import discord
from discord.ext import commands
import logging
from typing import Optional, List, Dict, Any, Union
import datetime
import re

logger = logging.getLogger("irlcord.discord_helpers")

def create_embed(
    title: str,
    description: str = None,
    color: int = 0x5865F2,  # Discord Blurple
    fields: List[Dict[str, Any]] = None,
    footer_text: str = None,
    thumbnail_url: str = None,
    image_url: str = None,
    author_name: str = None,
    author_icon_url: str = None,
    timestamp: bool = False
) -> discord.Embed:
    """Create a Discord embed with the given parameters"""
    embed = discord.Embed(
        title=title,
        description=description,
        color=color
    )
    
    if fields:
        for field in fields:
            embed.add_field(
                name=field["name"],
                value=field["value"],
                inline=field.get("inline", False)
            )
    
    if footer_text:
        embed.set_footer(text=footer_text)
    
    if thumbnail_url:
        embed.set_thumbnail(url=thumbnail_url)
    
    if image_url:
        embed.set_image(url=image_url)
    
    if author_name:
        embed.set_author(
            name=author_name,
            icon_url=author_icon_url
        )
    
    if timestamp:
        embed.timestamp = datetime.datetime.now()
    
    return embed

def create_event_embed(event: Dict[str, Any], attendees: List[Dict[str, Any]] = None, config: Dict[str, Any] = None) -> discord.Embed:
    """Create an embed for an event"""
    # Format date and time
    event_date = datetime.datetime.fromisoformat(event["date_time"])
    date_str = event_date.strftime("%A, %B %d, %Y")
    time_str = event_date.strftime("%I:%M %p")
    
    # Create description
    description = f"**Date:** {date_str}\n**Time:** {time_str}\n"
    
    if event.get("location_name"):
        description += f"**Location:** {event['location_name']}\n"
    
    if event.get("location_address"):
        description += f"**Address:** {event['location_address']}\n"
    
    if event.get("description"):
        description += f"\n{event['description']}\n"
    
    # Create fields for attendees
    fields = []
    
    if attendees:
        attending = [a for a in attendees if a["rsvp_status"] == "ATTENDING"]
        waitlist = [a for a in attendees if a["rsvp_status"] == "WAITLIST"]
        declined = [a for a in attendees if a["rsvp_status"] == "DECLINED"]
        
        if attending:
            attendee_names = "\n".join([f"<@{a['user_id']}>" for a in attending])
            fields.append({
                "name": f"Attending ({len(attending)})",
                "value": attendee_names if attendee_names else "No one yet",
                "inline": True
            })
        
        if waitlist:
            waitlist_names = "\n".join([f"<@{a['user_id']}>" for a in waitlist])
            fields.append({
                "name": f"Waitlist ({len(waitlist)})",
                "value": waitlist_names,
                "inline": True
            })
        
        if declined:
            declined_names = "\n".join([f"<@{a['user_id']}>" for a in declined])
            fields.append({
                "name": f"Declined ({len(declined)})",
                "value": declined_names,
                "inline": True
            })
    
    # Status indicator
    status_emoji = "🟢"  # Default: Approved
    if event["status"] == "pending":
        status_emoji = "🟠"
    elif event["status"] == "rejected":
        status_emoji = "🔴"
    
    # Create the embed
    embed = create_embed(
        title=f"{status_emoji} {event['name']}",
        description=description,
        fields=fields,
        footer_text=f"Event ID: {event['event_id']} • Host: <@{event['host_id']}>",
        timestamp=True
    )
    
    return embed

def create_group_embed(group: Dict[str, Any], members: List[Dict[str, Any]] = None, config: Dict[str, Any] = None) -> discord.Embed:
    """Create an embed for a group"""
    # Create description
    description = group.get("description", "No description provided.")
    
    # Add group settings
    settings = []
    
    if group.get("is_open") is not None:
        settings.append(f"Open Group: {'Yes' if group['is_open'] else 'No'}")
    
    if group.get("new_members_can_create_events") is not None:
        settings.append(f"New Members Can Create Events: {'Yes' if group['new_members_can_create_events'] else 'No'}")
    
    if group.get("event_approval_mode"):
        settings.append(f"Event Approval Mode: {group['event_approval_mode'].capitalize()}")
    
    if group.get("event_attendee_management_mode"):
        settings.append(f"Attendee Management: {group['event_attendee_management_mode'].capitalize()}")
    
    if group.get("contributor_events_required"):
        settings.append(f"Events Required for Contributor: {group['contributor_events_required']}")
    
    if settings:
        description += "\n\n**Settings:**\n" + "\n".join(settings)
    
    # Create fields for members
    fields = []
    
    if members:
        leaders = [m for m in members if m.get("is_leader")]
        regular_members = [m for m in members if not m.get("is_leader")]
        
        if leaders:
            leader_names = "\n".join([f"<@{m['user_id']}>" for m in leaders])
            fields.append({
                "name": "Leaders",
                "value": leader_names,
                "inline": True
            })
        
        if regular_members:
            member_names = "\n".join([f"<@{m['user_id']}>" for m in regular_members[:10]])
            if len(regular_members) > 10:
                member_names += f"\n... and {len(regular_members) - 10} more"
            
            fields.append({
                "name": f"Members ({len(regular_members)})",
                "value": member_names,
                "inline": True
            })
    
    # Create the embed
    embed = create_embed(
        title=group["name"],
        description=description,
        fields=fields,
        footer_text=f"Group ID: {group['group_id']} • Created: {group['created_at']}",
        timestamp=False
    )
    
    return embed

async def get_or_create_thread(
    channel: discord.TextChannel,
    name: str,
    message: Optional[discord.Message] = None,
    auto_archive_duration: int = 1440  # 1 day in minutes
) -> Optional[discord.Thread]:
    """Get an existing thread by name or create a new one"""
    # Try to find an existing thread with the same name
    for thread in channel.threads:
        if thread.name == name and not thread.archived:
            return thread
    
    # Create a new thread
    try:
        if message:
            thread = await message.create_thread(
                name=name,
                auto_archive_duration=auto_archive_duration
            )
        else:
            thread = await channel.create_thread(
                name=name,
                type=discord.ChannelType.public_thread,
                auto_archive_duration=auto_archive_duration
            )
        return thread
    except discord.HTTPException as e:
        logger.error(f"Failed to create thread: {e}")
        return None

def parse_command_args(content: str) -> Dict[str, str]:
    """Parse command arguments from a message content"""
    # Remove command prefix and command name
    parts = content.split(maxsplit=1)
    if len(parts) < 2:
        return {}
    
    args_text = parts[1]
    
    # Parse key-value pairs
    args = {}
    
    # Match quoted values first
    quoted_pattern = re.compile(r'(\w+)="([^"]*)"')
    for match in quoted_pattern.finditer(args_text):
        key, value = match.groups()
        args[key.lower()] = value
        args_text = args_text.replace(match.group(0), "")
    
    # Then match unquoted values
    unquoted_pattern = re.compile(r'(\w+)=(\S+)')
    for match in unquoted_pattern.finditer(args_text):
        key, value = match.groups()
        args[key.lower()] = value
    
    return args

def is_admin(ctx: commands.Context, admin_ids: List[str]) -> bool:
    """Check if a user is an admin"""
    return str(ctx.author.id) in admin_ids

def is_in_group_channel(ctx: commands.Context, db) -> bool:
    """Check if a command is used in a group channel"""
    group = db.fetch_one(
        "SELECT * FROM Groups WHERE channel_id = ?",
        (str(ctx.channel.id),)
    )
    return group is not None

def get_group_from_channel(ctx: commands.Context, db) -> Optional[Dict[str, Any]]:
    """Get a group from a channel"""
    return db.fetch_one(
        "SELECT * FROM Groups WHERE channel_id = ?",
        (str(ctx.channel.id),)
    )

def get_event_from_thread(ctx: commands.Context, db) -> Optional[Dict[str, Any]]:
    """Get an event from a thread"""
    if not isinstance(ctx.channel, discord.Thread):
        return None
    
    return db.fetch_one(
        "SELECT * FROM Events WHERE thread_id = ?",
        (str(ctx.channel.id),)
    ) 