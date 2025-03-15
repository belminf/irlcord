import sqlite3
import logging
from pathlib import Path
from typing import Dict, List, Any, Optional, Union

logger = logging.getLogger("irlcord.db")

class Database:
    def __init__(self, db_path: Path):
        self.db_path = db_path
        self.conn = None
        
    def connect(self):
        """Connect to the SQLite database"""
        try:
            self.conn = sqlite3.connect(self.db_path)
            self.conn.row_factory = sqlite3.Row
            logger.info(f"Connected to database at {self.db_path}")
            return True
        except sqlite3.Error as e:
            logger.error(f"Database connection error: {e}")
            return False
    
    def close(self):
        """Close the database connection"""
        if self.conn:
            self.conn.close()
            logger.info("Database connection closed")
    
    def execute(self, query: str, params: tuple = ()) -> Optional[sqlite3.Cursor]:
        """Execute a query with parameters"""
        try:
            cursor = self.conn.execute(query, params)
            self.conn.commit()
            return cursor
        except sqlite3.Error as e:
            logger.error(f"Database query error: {e}")
            logger.error(f"Query: {query}")
            logger.error(f"Params: {params}")
            return None
    
    def fetch_one(self, query: str, params: tuple = ()) -> Optional[Dict[str, Any]]:
        """Fetch a single row as a dictionary"""
        cursor = self.execute(query, params)
        if cursor:
            row = cursor.fetchone()
            return dict(row) if row else None
        return None
    
    def fetch_all(self, query: str, params: tuple = ()) -> List[Dict[str, Any]]:
        """Fetch all rows as a list of dictionaries"""
        cursor = self.execute(query, params)
        if cursor:
            return [dict(row) for row in cursor.fetchall()]
        return []
    
    # User operations
    def get_user(self, user_id: str) -> Optional[Dict[str, Any]]:
        """Get a user by Discord ID"""
        return self.fetch_one("SELECT * FROM Users WHERE user_id = ?", (user_id,))
    
    def create_user(self, user_id: str) -> bool:
        """Create a new user"""
        cursor = self.execute(
            "INSERT INTO Users (user_id) VALUES (?) ON CONFLICT DO NOTHING",
            (user_id,)
        )
        return cursor is not None
    
    def update_user(self, user_id: str, data: Dict[str, Any]) -> bool:
        """Update user information"""
        if not data:
            return False
        
        set_clause = ", ".join([f"{key} = ?" for key in data.keys()])
        params = list(data.values()) + [user_id]
        
        cursor = self.execute(
            f"UPDATE Users SET {set_clause} WHERE user_id = ?",
            tuple(params)
        )
        return cursor is not None
    
    # Group operations
    def get_group(self, group_id: int) -> Optional[Dict[str, Any]]:
        """Get a group by ID"""
        return self.fetch_one("SELECT * FROM Groups WHERE group_id = ?", (group_id,))
    
    def get_group_by_name(self, name: str) -> Optional[Dict[str, Any]]:
        """Get a group by name"""
        return self.fetch_one("SELECT * FROM Groups WHERE name = ?", (name,))
    
    def create_group(self, data: Dict[str, Any]) -> Optional[int]:
        """Create a new group"""
        keys = ", ".join(data.keys())
        placeholders = ", ".join(["?"] * len(data))
        
        cursor = self.execute(
            f"INSERT INTO Groups ({keys}) VALUES ({placeholders})",
            tuple(data.values())
        )
        
        if cursor:
            return cursor.lastrowid
        return None
    
    def update_group(self, group_id: int, data: Dict[str, Any]) -> bool:
        """Update group information"""
        if not data:
            return False
        
        set_clause = ", ".join([f"{key} = ?" for key in data.keys()])
        params = list(data.values()) + [group_id]
        
        cursor = self.execute(
            f"UPDATE Groups SET {set_clause} WHERE group_id = ?",
            tuple(params)
        )
        return cursor is not None
    
    # Group member operations
    def add_group_member(self, group_id: int, user_id: str, is_leader: bool = False) -> bool:
        """Add a user to a group"""
        cursor = self.execute(
            "INSERT INTO GroupMembers (group_id, user_id, is_leader) VALUES (?, ?, ?) ON CONFLICT DO NOTHING",
            (group_id, user_id, is_leader)
        )
        return cursor is not None
    
    def remove_group_member(self, group_id: int, user_id: str) -> bool:
        """Remove a user from a group"""
        cursor = self.execute(
            "DELETE FROM GroupMembers WHERE group_id = ? AND user_id = ?",
            (group_id, user_id)
        )
        return cursor is not None
    
    def is_group_member(self, group_id: int, user_id: str) -> bool:
        """Check if a user is a member of a group"""
        result = self.fetch_one(
            "SELECT 1 FROM GroupMembers WHERE group_id = ? AND user_id = ?",
            (group_id, user_id)
        )
        return result is not None
    
    def is_group_leader(self, group_id: int, user_id: str) -> bool:
        """Check if a user is a leader of a group"""
        result = self.fetch_one(
            "SELECT 1 FROM GroupMembers WHERE group_id = ? AND user_id = ? AND is_leader = 1",
            (group_id, user_id)
        )
        return result is not None
    
    # Event operations
    def create_event(self, data: Dict[str, Any]) -> Optional[int]:
        """Create a new event"""
        keys = ", ".join(data.keys())
        placeholders = ", ".join(["?"] * len(data))
        
        cursor = self.execute(
            f"INSERT INTO Events ({keys}) VALUES ({placeholders})",
            tuple(data.values())
        )
        
        if cursor:
            return cursor.lastrowid
        return None
    
    def get_event(self, event_id: int) -> Optional[Dict[str, Any]]:
        """Get an event by ID"""
        return self.fetch_one("SELECT * FROM Events WHERE event_id = ?", (event_id,))
    
    def update_event(self, event_id: int, data: Dict[str, Any]) -> bool:
        """Update event information"""
        if not data:
            return False
        
        set_clause = ", ".join([f"{key} = ?" for key in data.keys()])
        params = list(data.values()) + [event_id]
        
        cursor = self.execute(
            f"UPDATE Events SET {set_clause} WHERE event_id = ?",
            tuple(params)
        )
        return cursor is not None
    
    def get_group_events(self, group_id: int, status: str = None) -> List[Dict[str, Any]]:
        """Get all events for a group"""
        if status:
            return self.fetch_all(
                "SELECT * FROM Events WHERE group_id = ? AND status = ? ORDER BY date_time",
                (group_id, status)
            )
        return self.fetch_all(
            "SELECT * FROM Events WHERE group_id = ? ORDER BY date_time",
            (group_id,)
        )
    
    # Event attendee operations
    def add_event_attendee(self, event_id: int, user_id: str, rsvp_status: str = "ATTENDING") -> bool:
        """Add an attendee to an event"""
        cursor = self.execute(
            "INSERT INTO EventAttendees (event_id, user_id, rsvp_status) VALUES (?, ?, ?) "
            "ON CONFLICT(event_id, user_id) DO UPDATE SET rsvp_status = ?",
            (event_id, user_id, rsvp_status, rsvp_status)
        )
        return cursor is not None
    
    def remove_event_attendee(self, event_id: int, user_id: str) -> bool:
        """Remove an attendee from an event"""
        cursor = self.execute(
            "DELETE FROM EventAttendees WHERE event_id = ? AND user_id = ?",
            (event_id, user_id)
        )
        return cursor is not None
    
    def get_event_attendees(self, event_id: int, rsvp_status: str = None) -> List[Dict[str, Any]]:
        """Get all attendees for an event"""
        if rsvp_status:
            return self.fetch_all(
                "SELECT ea.*, u.venmo_username, u.dietary_restrictions, u.email "
                "FROM EventAttendees ea "
                "JOIN Users u ON ea.user_id = u.user_id "
                "WHERE ea.event_id = ? AND ea.rsvp_status = ?",
                (event_id, rsvp_status)
            )
        return self.fetch_all(
            "SELECT ea.*, u.venmo_username, u.dietary_restrictions, u.email "
            "FROM EventAttendees ea "
            "JOIN Users u ON ea.user_id = u.user_id "
            "WHERE ea.event_id = ?",
            (event_id,)
        )
    
    # Bill operations
    def create_bill(self, event_id: int, user_id: str, amount: float) -> Optional[int]:
        """Create a new bill"""
        cursor = self.execute(
            "INSERT INTO Bills (event_id, user_id, amount, paid) VALUES (?, ?, ?, 0)",
            (event_id, user_id, amount)
        )
        
        if cursor:
            return cursor.lastrowid
        return None
    
    def update_bill_status(self, bill_id: int, paid: bool) -> bool:
        """Update bill payment status"""
        cursor = self.execute(
            "UPDATE Bills SET paid = ? WHERE bill_id = ?",
            (1 if paid else 0, bill_id)
        )
        return cursor is not None
    
    def get_event_bills(self, event_id: int) -> List[Dict[str, Any]]:
        """Get all bills for an event"""
        return self.fetch_all(
            "SELECT b.*, u.venmo_username "
            "FROM Bills b "
            "JOIN Users u ON b.user_id = u.user_id "
            "WHERE b.event_id = ?",
            (event_id,)
        ) 