import discord
from discord.ext import commands
import logging
from typing import Optional, List, Dict, Any
import asyncio
import datetime
import re

from irlcord.utils.discord_helpers import (
    create_embed,
    create_event_embed,
    parse_command_args,
    is_admin,
    is_in_group_channel,
    get_group_from_channel,
    get_event_from_thread,
    get_or_create_thread
)

logger = logging.getLogger("irlcord.cogs.events")

class Events(commands.Cog):
    """Commands for managing events"""
    
    def __init__(self, bot):
        self.bot = bot
        self.db = bot.db
        self.config = bot.config
        self.terminology = bot.config.get("terminology", {})
        self.commands_config = bot.config.get("commands", {})
    
    @property
    def group_singular(self):
        return self.terminology.get("group_singular", "Circle")
    
    @property
    def event_singular(self):
        return self.terminology.get("event_singular", "Event")
    
    @property
    def event_plural(self):
        return self.terminology.get("event_plural", "Events")
    
    async def get_event_attendees(self, event_id: int) -> List[Dict[str, Any]]:
        """Get all attendees for an event"""
        return self.db.get_event_attendees(event_id)
    
    @commands.Cog.listener()
    async def on_message(self, message):
        """Listen for event commands in messages"""
        if message.author.bot or not message.content:
            return
        
        # Check if the message starts with an event command
        event_create_cmd = self.commands_config.get("event_create", "event new")
        event_modify_cmd = self.commands_config.get("event_modify", "event modify")
        event_confirm_cmd = self.commands_config.get("event_confirm", "event confirm")
        event_unconfirm_cmd = self.commands_config.get("event_unconfirm", "event unconfirm")
        event_waitlist_cmd = self.commands_config.get("event_waitlist", "event waitlist")
        event_info_cmd = self.commands_config.get("event_info", "event info")
        event_change_host_cmd = self.commands_config.get("event_change_host", "event change host")
        
        content_lower = message.content.lower()
        
        # Process commands
        ctx = await self.bot.get_context(message)
        if content_lower.startswith(event_create_cmd):
            await self.event_create(ctx)
        elif content_lower.startswith(event_modify_cmd):
            await self.event_modify(ctx)
        elif content_lower.startswith(event_confirm_cmd):
            await self.event_confirm(ctx)
        elif content_lower.startswith(event_unconfirm_cmd):
            await self.event_unconfirm(ctx)
        elif content_lower.startswith(event_waitlist_cmd):
            await self.event_waitlist(ctx)
        elif content_lower.startswith(event_info_cmd):
            await self.event_info(ctx)
        elif content_lower.startswith(event_change_host_cmd):
            await self.event_change_host(ctx)
    
    async def event_create(self, ctx):
        """Create a new event"""
        # Check if in a group channel
        group = get_group_from_channel(ctx, self.db)
        if not group:
            await ctx.send(f"This command must be used in a {self.group_singular.lower()} channel.")
            return
        
        # Check if user can create events
        if not group["new_members_can_create_events"] and not self.db.is_group_leader(group["group_id"], str(ctx.author.id)):
            await ctx.send(f"Only leaders can create events in this {self.group_singular.lower()}.")
            return
        
        # Parse arguments
        args = parse_command_args(ctx.message.content)
        
        if not args.get("name"):
            await ctx.send(f"Please provide a name for the {self.event_singular.lower()}. Example: `event new name=\"Game Night\" date=\"2024-03-20\" time=\"19:00\"`")
            return
        
        if not args.get("date") or not args.get("time"):
            await ctx.send("Please provide both date (YYYY-MM-DD) and time (HH:MM) for the event.")
            return
        
        try:
            # Parse date and time
            date_str = args["date"]
            time_str = args["time"]
            date_time = datetime.datetime.strptime(f"{date_str} {time_str}", "%Y-%m-%d %H:%M")
            
            if date_time < datetime.datetime.now():
                await ctx.send("Event date and time must be in the future.")
                return
            
        except ValueError:
            await ctx.send("Invalid date or time format. Please use YYYY-MM-DD for date and HH:MM for time.")
            return
        
        # Create event data
        event_data = {
            "group_id": group["group_id"],
            "host_id": str(ctx.author.id),
            "name": args["name"],
            "date_time": date_time.isoformat(),
            "location_name": args.get("location", ""),
            "location_address": args.get("address", ""),
            "description": args.get("description", ""),
            "max_attendees": int(args["max"]) if args.get("max") else None,
            "is_public": args.get("public", "true").lower() == "true",
            "status": "pending" if group["event_approval_mode"] != "none" else "approved"
        }
        
        # Create the event
        event_id = self.db.create_event(event_data)
        
        if not event_id:
            await ctx.send(f"Failed to create {self.event_singular.lower()}.")
            return
        
        # Get the created event
        event = self.db.get_event(event_id)
        
        # Create thread for the event
        thread_name = f"{event['name']}-{date_time.strftime('%Y-%m-%d')}"
        thread = await get_or_create_thread(ctx.channel, thread_name)
        
        if not thread:
            await ctx.send(f"Failed to create thread for {self.event_singular.lower()}.")
            return
        
        # Update event with thread ID
        self.db.update_event(event_id, {"thread_id": str(thread.id)})
        
        # Add host as an attendee
        self.db.add_event_attendee(event_id, str(ctx.author.id))
        
        # Get attendees
        attendees = await self.get_event_attendees(event_id)
        
        # Create embed
        embed = create_event_embed(event, attendees, self.config)
        
        # Send initial message
        status_msg = ""
        if event["status"] == "pending":
            status_msg = f"\n\nThis {self.event_singular.lower()} requires approval from a leader."
        
        await thread.send(
            f"**{event['name']}** has been created by <@{ctx.author.id}>!{status_msg}",
            embed=embed
        )
        
        await ctx.send(f"{self.event_singular} created! Check out <#{thread.id}>")
    
    async def event_modify(self, ctx):
        """Modify an event"""
        # Check if in an event thread
        event = get_event_from_thread(ctx, self.db)
        if not event:
            await ctx.send(f"This command must be used in an {self.event_singular.lower()} thread.")
            return
        
        # Check if user is host or leader
        group = self.db.get_group(event["group_id"])
        if str(ctx.author.id) != event["host_id"] and not self.db.is_group_leader(group["group_id"], str(ctx.author.id)):
            await ctx.send(f"Only the host or a leader can modify this {self.event_singular.lower()}.")
            return
        
        # Parse arguments
        args = parse_command_args(ctx.message.content)
        
        if not args:
            await ctx.send(f"Please provide settings to modify. Example: `event modify description=\"New description\"`")
            return
        
        # Prepare data for update
        update_data = {}
        
        if "name" in args:
            update_data["name"] = args["name"]
        
        if "description" in args:
            update_data["description"] = args["description"]
        
        if "location" in args:
            update_data["location_name"] = args["location"]
        
        if "address" in args:
            update_data["location_address"] = args["address"]
        
        if "max" in args:
            try:
                update_data["max_attendees"] = int(args["max"])
            except ValueError:
                await ctx.send("Maximum attendees must be a number.")
                return
        
        if "date" in args or "time" in args:
            try:
                current_date = datetime.datetime.fromisoformat(event["date_time"])
                
                if "date" in args:
                    date_str = args["date"]
                    new_date = datetime.datetime.strptime(date_str, "%Y-%m-%d")
                    current_date = current_date.replace(
                        year=new_date.year,
                        month=new_date.month,
                        day=new_date.day
                    )
                
                if "time" in args:
                    time_str = args["time"]
                    new_time = datetime.datetime.strptime(time_str, "%H:%M")
                    current_date = current_date.replace(
                        hour=new_time.hour,
                        minute=new_time.minute
                    )
                
                if current_date < datetime.datetime.now():
                    await ctx.send("Event date and time must be in the future.")
                    return
                
                update_data["date_time"] = current_date.isoformat()
                
            except ValueError:
                await ctx.send("Invalid date or time format. Please use YYYY-MM-DD for date and HH:MM for time.")
                return
        
        # Update the event
        if update_data:
            success = self.db.update_event(event["event_id"], update_data)
            
            if success:
                # Update thread name if name changed
                if "name" in update_data:
                    try:
                        new_thread_name = f"{update_data['name']}-{datetime.datetime.fromisoformat(event['date_time']).strftime('%Y-%m-%d')}"
                        await ctx.channel.edit(name=new_thread_name)
                    except discord.HTTPException as e:
                        logger.error(f"Failed to update thread name: {e}")
                
                # Get updated event and attendees
                updated_event = self.db.get_event(event["event_id"])
                attendees = await self.get_event_attendees(event["event_id"])
                
                # Create embed
                embed = create_event_embed(updated_event, attendees, self.config)
                
                await ctx.send(f"{self.event_singular} updated!", embed=embed)
            else:
                await ctx.send(f"Failed to update {self.event_singular.lower()}.")
        else:
            await ctx.send("No valid settings provided to update.")
    
    async def event_confirm(self, ctx):
        """Confirm attendance for an event"""
        # Check if in an event thread
        event = get_event_from_thread(ctx, self.db)
        if not event:
            await ctx.send(f"This command must be used in an {self.event_singular.lower()} thread.")
            return
        
        # Check if event is approved
        if event["status"] != "approved":
            await ctx.send(f"This {self.event_singular.lower()} has not been approved yet.")
            return
        
        # Check if user is in the group
        group = self.db.get_group(event["group_id"])
        if not self.db.is_group_member(group["group_id"], str(ctx.author.id)):
            await ctx.send(f"You must be a member of the {self.group_singular.lower()} to attend events.")
            return
        
        # Get current attendees
        attendees = await self.get_event_attendees(event["event_id"])
        attending = [a for a in attendees if a["rsvp_status"] == "ATTENDING"]
        
        # Check if event is full
        if event["max_attendees"] and len(attending) >= event["max_attendees"]:
            # Add to waitlist
            self.db.add_event_attendee(event["event_id"], str(ctx.author.id), "WAITLIST")
            await ctx.send(f"The {self.event_singular.lower()} is full. You have been added to the waitlist.")
        else:
            # Add as attending
            self.db.add_event_attendee(event["event_id"], str(ctx.author.id), "ATTENDING")
            await ctx.send(f"You are now attending this {self.event_singular.lower()}!")
        
        # Update event embed
        updated_attendees = await self.get_event_attendees(event["event_id"])
        embed = create_event_embed(event, updated_attendees, self.config)
        await ctx.send(embed=embed)
    
    async def event_unconfirm(self, ctx):
        """Remove attendance confirmation for an event"""
        # Check if in an event thread
        event = get_event_from_thread(ctx, self.db)
        if not event:
            await ctx.send(f"This command must be used in an {self.event_singular.lower()} thread.")
            return
        
        # Remove attendance
        self.db.remove_event_attendee(event["event_id"], str(ctx.author.id))
        
        # Move someone from waitlist if available
        attendees = await self.get_event_attendees(event["event_id"])
        waitlist = [a for a in attendees if a["rsvp_status"] == "WAITLIST"]
        
        if waitlist:
            # Move first person from waitlist to attending
            self.db.add_event_attendee(event["event_id"], waitlist[0]["user_id"], "ATTENDING")
            await ctx.send(f"<@{waitlist[0]['user_id']}> has been moved from the waitlist to attending!")
        
        await ctx.send(f"You are no longer attending this {self.event_singular.lower()}.")
        
        # Update event embed
        updated_attendees = await self.get_event_attendees(event["event_id"])
        embed = create_event_embed(event, updated_attendees, self.config)
        await ctx.send(embed=embed)
    
    async def event_waitlist(self, ctx):
        """Join the waitlist for an event"""
        # Check if in an event thread
        event = get_event_from_thread(ctx, self.db)
        if not event:
            await ctx.send(f"This command must be used in an {self.event_singular.lower()} thread.")
            return
        
        # Check if event is approved
        if event["status"] != "approved":
            await ctx.send(f"This {self.event_singular.lower()} has not been approved yet.")
            return
        
        # Check if user is in the group
        group = self.db.get_group(event["group_id"])
        if not self.db.is_group_member(group["group_id"], str(ctx.author.id)):
            await ctx.send(f"You must be a member of the {self.group_singular.lower()} to join the waitlist.")
            return
        
        # Add to waitlist
        self.db.add_event_attendee(event["event_id"], str(ctx.author.id), "WAITLIST")
        await ctx.send(f"You have been added to the waitlist for this {self.event_singular.lower()}.")
        
        # Update event embed
        attendees = await self.get_event_attendees(event["event_id"])
        embed = create_event_embed(event, attendees, self.config)
        await ctx.send(embed=embed)
    
    async def event_info(self, ctx):
        """Show information about an event"""
        # Check if in an event thread or if ID is provided
        args = parse_command_args(ctx.message.content)
        
        if args.get("id"):
            # Get event by ID
            try:
                event_id = int(args["id"])
                event = self.db.get_event(event_id)
            except ValueError:
                await ctx.send("Invalid event ID.")
                return
        else:
            # Get event from thread
            event = get_event_from_thread(ctx, self.db)
        
        if not event:
            await ctx.send(f"No {self.event_singular.lower()} found.")
            return
        
        # Get attendees
        attendees = await self.get_event_attendees(event["event_id"])
        
        # Create embed
        embed = create_event_embed(event, attendees, self.config)
        
        await ctx.send(embed=embed)
    
    async def event_change_host(self, ctx):
        """Change the host of an event"""
        # Check if in an event thread
        event = get_event_from_thread(ctx, self.db)
        if not event:
            await ctx.send(f"This command must be used in an {self.event_singular.lower()} thread.")
            return
        
        # Check if user is current host or leader
        group = self.db.get_group(event["group_id"])
        if str(ctx.author.id) != event["host_id"] and not self.db.is_group_leader(group["group_id"], str(ctx.author.id)):
            await ctx.send(f"Only the current host or a leader can change the host.")
            return
        
        # Parse arguments
        args = parse_command_args(ctx.message.content)
        
        if not args.get("user"):
            await ctx.send("Please mention the new host. Example: `event change host user=@username`")
            return
        
        # Extract user ID from mention
        user_id = re.findall(r"\d+", args["user"])[0]
        
        # Check if new host is in the group
        if not self.db.is_group_member(group["group_id"], user_id):
            await ctx.send(f"The new host must be a member of the {self.group_singular.lower()}.")
            return
        
        # Update host
        success = self.db.update_event(event["event_id"], {"host_id": user_id})
        
        if success:
            # Get updated event and attendees
            updated_event = self.db.get_event(event["event_id"])
            attendees = await self.get_event_attendees(event["event_id"])
            
            # Create embed
            embed = create_event_embed(updated_event, attendees, self.config)
            
            await ctx.send(f"Event host changed to <@{user_id}>!", embed=embed)
        else:
            await ctx.send("Failed to change event host.")

async def setup(bot):
    await bot.add_cog(Events(bot)) 