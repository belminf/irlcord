package models

import (
	"time"
)

// User represents a Discord user
type User struct {
	UserID             string    `db:"user_id"`
	VenmoUsername      string    `db:"venmo_username"`
	DietaryRestrictions string    `db:"dietary_restrictions"`
	Email              string    `db:"email"`
	JoinedAt           time.Time `db:"joined_at"`
}

// Group represents a group in the database
type Group struct {
	GroupID                    int64     `json:"group_id"`
	Name                       string    `json:"name"`
	Description                string    `json:"description"`
	ChannelID                  string    `json:"channel_id"`
	IsOpen                     bool      `json:"is_open"`
	NewMembersCanCreateEvents  bool      `json:"new_members_can_create_events"`
	EventApprovalMode          string    `json:"event_approval_mode"`
	EventAttendeeManagementMode string    `json:"event_attendee_management_mode"`
	ContributorEventsRequired  int       `json:"contributor_events_required"`
	CreatedAt                  time.Time `json:"created_at"`
	UpdatedAt                  time.Time `json:"updated_at"`
}

// GroupMember represents a member of a group
type GroupMember struct {
	GroupID       int64     `json:"group_id"`
	UserID        string    `json:"user_id"`
	IsLeader      bool      `json:"is_leader"`
	IsContributor bool      `json:"is_contributor"`
	JoinedAt      time.Time `json:"joined_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// Event represents an event in the database
type Event struct {
	EventID         int64     `json:"event_id"`
	GroupID         int64     `json:"group_id"`
	HostID          string    `json:"host_id"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	DateTime        time.Time `json:"date_time"`
	LocationName    string    `json:"location_name"`
	LocationAddress string    `json:"location_address"`
	MaxAttendees    int       `json:"max_attendees"`
	Status          string    `json:"status"`
	MessageID       string    `json:"message_id"`
	ThreadID        string    `json:"thread_id"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// EventAttendee represents an attendee of an event
type EventAttendee struct {
	EventID    int64     `json:"event_id"`
	UserID     string    `json:"user_id"`
	RSVPStatus string    `json:"rsvp_status"`
	RSVPTime   time.Time `json:"rsvp_time"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Bill represents a bill for an event
type Bill struct {
	BillID  int64   `db:"bill_id"`
	EventID int64   `db:"event_id"`
	UserID  string  `db:"user_id"`
	Amount  float64 `db:"amount"`
	Paid    bool    `db:"paid"`
}

// Settings represents global settings for the bot
type Settings struct {
	GuildID      string            `json:"guild_id"`
	Terminology  map[string]string `json:"terminology"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

// EventStatus represents the status of an event
type EventStatus string

// RSVPStatus represents the RSVP status of an attendee
type RSVPStatus string

// Event status constants
const (
	EventStatusPending  EventStatus = "pending"
	EventStatusApproved EventStatus = "approved"
	EventStatusRejected EventStatus = "rejected"
	EventStatusCanceled EventStatus = "canceled"
)

// RSVP status constants
const (
	RSVPStatusAttending RSVPStatus = "attending"
	RSVPStatusWaitlist  RSVPStatus = "waitlist"
	RSVPStatusDeclined  RSVPStatus = "declined"
)

// Event approval mode constants
const (
	EventApprovalModeNone     = "none"
	EventApprovalModeLeaders  = "leaders"
	EventApprovalModeManual   = "manual"
)

// Event attendee management mode constants
const (
	EventAttendeeManagementModeOpen     = "open"
	EventAttendeeManagementModeLeaders  = "leaders"
	EventAttendeeManagementModeHost     = "host"
) 