import discord
from discord.ext import commands
import logging
from typing import Optional, List, Dict, Any
import asyncio

from irlcord.utils.discord_helpers import (
    create_embed,
    create_group_embed,
    parse_command_args,
    is_admin,
    is_in_group_channel,
    get_group_from_channel
)

logger = logging.getLogger("irlcord.cogs.groups")

class Groups(commands.Cog):
    """Commands for managing groups/circles"""
    
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
    def group_plural(self):
        return self.terminology.get("group_plural", "Circles")
    
    @property
    def leader_singular(self):
        return self.terminology.get("leader_singular", "Leader")
    
    @property
    def contributor_singular(self):
        return self.terminology.get("contributor_singular", "Adventurer")
    
    async def get_group_members(self, group_id: int) -> List[Dict[str, Any]]:
        """Get all members of a group"""
        return self.db.fetch_all(
            "SELECT gm.*, u.venmo_username, u.dietary_restrictions, u.email "
            "FROM GroupMembers gm "
            "JOIN Users u ON gm.user_id = u.user_id "
            "WHERE gm.group_id = ?",
            (group_id,)
        )
    
    @commands.Cog.listener()
    async def on_message(self, message):
        """Listen for group commands in messages"""
        if message.author.bot or not message.content:
            return
        
        # Check if the message starts with a group command
        group_create_cmd = self.commands_config.get("group_create", "circle new")
        group_join_cmd = self.commands_config.get("group_join", "circle join")
        group_leave_cmd = self.commands_config.get("group_leave", "circle leave")
        group_info_cmd = self.commands_config.get("group_info", "circle info")
        group_modify_cmd = self.commands_config.get("group_modify", "circle modify")
        
        content_lower = message.content.lower()
        
        # Process commands
        ctx = await self.bot.get_context(message)
        if content_lower.startswith(group_create_cmd):
            await self.group_create(ctx)
        elif content_lower.startswith(group_join_cmd):
            await self.group_join(ctx)
        elif content_lower.startswith(group_leave_cmd):
            await self.group_leave(ctx)
        elif content_lower.startswith(group_info_cmd):
            await self.group_info(ctx)
        elif content_lower.startswith(group_modify_cmd):
            await self.group_modify(ctx)
    
    async def group_create(self, ctx):
        """Create a new group"""
        # Check if user is an admin
        admin_ids = self.config.get("general", {}).get("admin_user_ids", [])
        if not is_admin(ctx, admin_ids):
            await ctx.send("Only administrators can create new groups.")
            return
        
        # Parse arguments
        args = parse_command_args(ctx.message.content)
        
        if not args.get("name"):
            await ctx.send(f"Please provide a name for the {self.group_singular.lower()}. Example: `circle new name=\"Hiking Group\"`")
            return
        
        # Check if group already exists
        existing_group = self.db.get_group_by_name(args["name"])
        if existing_group:
            await ctx.send(f"A {self.group_singular.lower()} with that name already exists.")
            return
        
        # Create a new channel for the group
        try:
            channel = await ctx.guild.create_text_channel(
                name=args["name"].lower().replace(" ", "-"),
                topic=args.get("description", f"Channel for {args['name']}")
            )
            
            # Set permissions for the channel
            await channel.set_permissions(
                ctx.guild.default_role,
                read_messages=False,
                send_messages=False
            )
            
            # Create the group in the database
            group_data = {
                "name": args["name"],
                "description": args.get("description", ""),
                "is_open": args.get("is_open", "true").lower() == "true",
                "channel_id": str(channel.id),
                "new_members_can_create_events": args.get("new_members_can_create_events", "true").lower() == "true",
                "event_approval_mode": args.get("event_approval_mode", "public"),
                "event_attendee_management_mode": args.get("event_attendee_management_mode", "host")
            }
            
            if args.get("contributor_events_required"):
                group_data["contributor_events_required"] = int(args["contributor_events_required"])
            
            group_id = self.db.create_group(group_data)
            
            if not group_id:
                await ctx.send(f"Failed to create {self.group_singular.lower()} in the database.")
                await channel.delete()
                return
            
            # Add the creator as a leader
            self.db.add_group_member(group_id, str(ctx.author.id), is_leader=True)
            
            # Create welcome message
            group = self.db.get_group(group_id)
            members = await self.get_group_members(group_id)
            
            embed = create_group_embed(group, members, self.config)
            welcome_msg = await channel.send(
                f"Welcome to the **{args['name']}** {self.group_singular.lower()}! "
                f"This channel will be used for organizing events and discussions.",
                embed=embed
            )
            
            # Pin the welcome message
            await welcome_msg.pin()
            
            # Add creator to the channel
            await channel.set_permissions(ctx.author, read_messages=True, send_messages=True)
            
            await ctx.send(f"{self.group_singular} created successfully! Check out <#{channel.id}>")
            
        except discord.HTTPException as e:
            logger.error(f"Failed to create group channel: {e}")
            await ctx.send(f"Failed to create {self.group_singular.lower()} channel: {e}")
    
    async def group_join(self, ctx):
        """Join a group"""
        # Parse arguments
        args = parse_command_args(ctx.message.content)
        
        if not args.get("name"):
            await ctx.send(f"Please provide the name of the {self.group_singular.lower()} to join. Example: `circle join name=\"Hiking Group\"`")
            return
        
        # Get the group
        group = self.db.get_group_by_name(args["name"])
        if not group:
            await ctx.send(f"No {self.group_singular.lower()} found with that name.")
            return
        
        # Check if the group is open
        if not group["is_open"]:
            await ctx.send(f"This {self.group_singular.lower()} is not open for new members. Please contact a {self.leader_singular.lower()} to join.")
            return
        
        # Check if already a member
        if self.db.is_group_member(group["group_id"], str(ctx.author.id)):
            await ctx.send(f"You are already a member of this {self.group_singular.lower()}.")
            return
        
        # Add user to the group
        self.db.add_group_member(group["group_id"], str(ctx.author.id))
        
        # Add user to the channel
        channel = ctx.guild.get_channel(int(group["channel_id"]))
        if channel:
            await channel.set_permissions(ctx.author, read_messages=True, send_messages=True)
        
        # Create or get user in database
        self.db.create_user(str(ctx.author.id))
        
        await ctx.send(f"You have joined the **{group['name']}** {self.group_singular.lower()}! Check out <#{group['channel_id']}>")
        
        # Notify the group channel
        if channel:
            await channel.send(f"Welcome <@{ctx.author.id}> to the group!")
    
    async def group_leave(self, ctx):
        """Leave a group"""
        # Check if in a group channel
        group = get_group_from_channel(ctx, self.db)
        if not group:
            await ctx.send(f"This command must be used in a {self.group_singular.lower()} channel.")
            return
        
        # Check if a member
        if not self.db.is_group_member(group["group_id"], str(ctx.author.id)):
            await ctx.send(f"You are not a member of this {self.group_singular.lower()}.")
            return
        
        # Check if the last leader
        if self.db.is_group_leader(group["group_id"], str(ctx.author.id)):
            leaders = self.db.fetch_all(
                "SELECT * FROM GroupMembers WHERE group_id = ? AND is_leader = 1",
                (group["group_id"],)
            )
            
            if len(leaders) == 1:
                await ctx.send(f"You are the last {self.leader_singular.lower()} of this {self.group_singular.lower()}. "
                              f"Please assign another {self.leader_singular.lower()} before leaving.")
                return
        
        # Remove user from the group
        self.db.remove_group_member(group["group_id"], str(ctx.author.id))
        
        # Remove user from the channel
        await ctx.channel.set_permissions(ctx.author, overwrite=None)
        
        await ctx.send(f"You have left the **{group['name']}** {self.group_singular.lower()}.")
    
    async def group_info(self, ctx):
        """Show information about a group"""
        # Check if in a group channel or if name is provided
        args = parse_command_args(ctx.message.content)
        
        if args.get("name"):
            # Get group by name
            group = self.db.get_group_by_name(args["name"])
            if not group:
                await ctx.send(f"No {self.group_singular.lower()} found with that name.")
                return
        else:
            # Get group from channel
            group = get_group_from_channel(ctx, self.db)
            if not group:
                await ctx.send(f"This command must be used in a {self.group_singular.lower()} channel or with a name parameter.")
                return
        
        # Get members
        members = await self.get_group_members(group["group_id"])
        
        # Create embed
        embed = create_group_embed(group, members, self.config)
        
        await ctx.send(embed=embed)
    
    async def group_modify(self, ctx):
        """Modify group settings"""
        # Check if in a group channel
        group = get_group_from_channel(ctx, self.db)
        if not group:
            await ctx.send(f"This command must be used in a {self.group_singular.lower()} channel.")
            return
        
        # Check if a leader
        if not self.db.is_group_leader(group["group_id"], str(ctx.author.id)):
            await ctx.send(f"Only {self.leader_singular.lower()}s can modify {self.group_singular.lower()} settings.")
            return
        
        # Parse arguments
        args = parse_command_args(ctx.message.content)
        
        if not args:
            await ctx.send(f"Please provide settings to modify. Example: `circle modify is_open=false`")
            return
        
        # Prepare data for update
        update_data = {}
        
        if "name" in args:
            update_data["name"] = args["name"]
        
        if "description" in args:
            update_data["description"] = args["description"]
        
        if "is_open" in args:
            update_data["is_open"] = args["is_open"].lower() == "true"
        
        if "new_members_can_create_events" in args:
            update_data["new_members_can_create_events"] = args["new_members_can_create_events"].lower() == "true"
        
        if "event_approval_mode" in args:
            valid_modes = ["none", "public", "all"]
            if args["event_approval_mode"].lower() in valid_modes:
                update_data["event_approval_mode"] = args["event_approval_mode"].lower()
            else:
                await ctx.send(f"Invalid event approval mode. Valid options are: {', '.join(valid_modes)}")
                return
        
        if "event_attendee_management_mode" in args:
            valid_modes = ["host", "self"]
            if args["event_attendee_management_mode"].lower() in valid_modes:
                update_data["event_attendee_management_mode"] = args["event_attendee_management_mode"].lower()
            else:
                await ctx.send(f"Invalid attendee management mode. Valid options are: {', '.join(valid_modes)}")
                return
        
        if "contributor_events_required" in args:
            try:
                update_data["contributor_events_required"] = int(args["contributor_events_required"])
            except ValueError:
                await ctx.send("Contributor events required must be a number.")
                return
        
        # Update the group
        if update_data:
            success = self.db.update_group(group["group_id"], update_data)
            
            if success:
                # Update channel name if name changed
                if "name" in update_data:
                    try:
                        await ctx.channel.edit(
                            name=update_data["name"].lower().replace(" ", "-")
                        )
                    except discord.HTTPException as e:
                        logger.error(f"Failed to update channel name: {e}")
                
                # Update channel topic if description changed
                if "description" in update_data:
                    try:
                        await ctx.channel.edit(
                            topic=update_data["description"]
                        )
                    except discord.HTTPException as e:
                        logger.error(f"Failed to update channel topic: {e}")
                
                # Get updated group and members
                updated_group = self.db.get_group(group["group_id"])
                members = await self.get_group_members(group["group_id"])
                
                # Create embed
                embed = create_group_embed(updated_group, members, self.config)
                
                await ctx.send(f"{self.group_singular} settings updated!", embed=embed)
            else:
                await ctx.send(f"Failed to update {self.group_singular.lower()} settings.")
        else:
            await ctx.send("No valid settings provided to update.")
    
    @commands.Cog.listener()
    async def on_member_remove(self, member):
        """Remove user from all groups when they leave the server"""
        # Get all groups the user is a member of
        groups = self.db.fetch_all(
            "SELECT g.* FROM Groups g "
            "JOIN GroupMembers gm ON g.group_id = gm.group_id "
            "WHERE gm.user_id = ?",
            (str(member.id),)
        )
        
        # Remove user from all groups
        for group in groups:
            self.db.remove_group_member(group["group_id"], str(member.id))
            
            # Notify the group channel
            channel = member.guild.get_channel(int(group["channel_id"]))
            if channel:
                await channel.send(f"<@{member.id}> has left the server and has been removed from this {self.group_singular.lower()}.")

async def setup(bot):
    await bot.add_cog(Groups(bot)) 