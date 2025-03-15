package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/azlyth/irlcord/pkg/models"
	_ "github.com/mattn/go-sqlite3"
)

// Database represents a database connection
type Database struct {
	db *sql.DB
}

// New creates a new database connection
func New(path string) (*Database, error) {
	// Open the database
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	// Set connection parameters
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(time.Hour)

	// Create a new database instance
	database := &Database{
		db: db,
	}

	// Initialize the database
	err = database.initialize()
	if err != nil {
		return nil, fmt.Errorf("error initializing database: %w", err)
	}

	return database, nil
}

// Close closes the database connection
func (d *Database) Close() error {
	return d.db.Close()
}

// initialize creates the database tables if they don't exist
func (d *Database) initialize() error {
	// Create the groups table
	_, err := d.db.Exec(`
		CREATE TABLE IF NOT EXISTS groups (
			group_id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			description TEXT,
			channel_id TEXT,
			is_open BOOLEAN DEFAULT TRUE,
			new_members_can_create_events BOOLEAN DEFAULT TRUE,
			event_approval_mode TEXT DEFAULT 'none',
			event_attendee_management_mode TEXT DEFAULT 'open',
			contributor_events_required INTEGER DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("error creating groups table: %w", err)
	}

	// Create the group_members table
	_, err = d.db.Exec(`
		CREATE TABLE IF NOT EXISTS group_members (
			group_id INTEGER NOT NULL,
			user_id TEXT NOT NULL,
			is_leader BOOLEAN DEFAULT FALSE,
			is_contributor BOOLEAN DEFAULT FALSE,
			joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (group_id, user_id),
			FOREIGN KEY (group_id) REFERENCES groups (group_id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return fmt.Errorf("error creating group_members table: %w", err)
	}

	// Create the events table
	_, err = d.db.Exec(`
		CREATE TABLE IF NOT EXISTS events (
			event_id INTEGER PRIMARY KEY AUTOINCREMENT,
			group_id INTEGER NOT NULL,
			host_id TEXT NOT NULL,
			name TEXT NOT NULL,
			description TEXT,
			date_time TIMESTAMP NOT NULL,
			location_name TEXT,
			location_address TEXT,
			max_attendees INTEGER DEFAULT 0,
			status TEXT DEFAULT 'pending',
			message_id TEXT,
			thread_id TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (group_id) REFERENCES groups (group_id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return fmt.Errorf("error creating events table: %w", err)
	}

	// Create the event_attendees table
	_, err = d.db.Exec(`
		CREATE TABLE IF NOT EXISTS event_attendees (
			event_id INTEGER NOT NULL,
			user_id TEXT NOT NULL,
			rsvp_status TEXT NOT NULL,
			rsvp_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (event_id, user_id),
			FOREIGN KEY (event_id) REFERENCES events (event_id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return fmt.Errorf("error creating event_attendees table: %w", err)
	}

	// Create the settings table
	_, err = d.db.Exec(`
		CREATE TABLE IF NOT EXISTS settings (
			guild_id TEXT PRIMARY KEY,
			terminology TEXT,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("error creating settings table: %w", err)
	}

	return nil
}

// Group methods

// CreateGroup creates a new group
func (d *Database) CreateGroup(group *models.Group) (int64, error) {
	// Insert the group
	result, err := d.db.Exec(`
		INSERT INTO groups (
			name, description, channel_id, is_open, new_members_can_create_events,
			event_approval_mode, event_attendee_management_mode, contributor_events_required
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`,
		group.Name, group.Description, group.ChannelID, group.IsOpen, group.NewMembersCanCreateEvents,
		group.EventApprovalMode, group.EventAttendeeManagementMode, group.ContributorEventsRequired,
	)
	if err != nil {
		return 0, fmt.Errorf("error creating group: %w", err)
	}

	// Get the group ID
	groupID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("error getting group ID: %w", err)
	}

	return groupID, nil
}

// GetGroup gets a group by ID
func (d *Database) GetGroup(groupID int64) (*models.Group, error) {
	// Query the group
	row := d.db.QueryRow(`
		SELECT
			group_id, name, description, channel_id, is_open, new_members_can_create_events,
			event_approval_mode, event_attendee_management_mode, contributor_events_required,
			created_at, updated_at
		FROM groups
		WHERE group_id = ?
	`, groupID)

	// Scan the result
	var group models.Group
	err := row.Scan(
		&group.GroupID, &group.Name, &group.Description, &group.ChannelID, &group.IsOpen, &group.NewMembersCanCreateEvents,
		&group.EventApprovalMode, &group.EventAttendeeManagementMode, &group.ContributorEventsRequired,
		&group.CreatedAt, &group.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("error getting group: %w", err)
	}

	return &group, nil
}

// GetGroups gets all groups
func (d *Database) GetGroups() ([]*models.Group, error) {
	// Query the groups
	rows, err := d.db.Query(`
		SELECT
			group_id, name, description, channel_id, is_open, new_members_can_create_events,
			event_approval_mode, event_attendee_management_mode, contributor_events_required,
			created_at, updated_at
		FROM groups
		ORDER BY name
	`)
	if err != nil {
		return nil, fmt.Errorf("error getting groups: %w", err)
	}
	defer rows.Close()

	// Scan the results
	var groups []*models.Group
	for rows.Next() {
		var group models.Group
		err := rows.Scan(
			&group.GroupID, &group.Name, &group.Description, &group.ChannelID, &group.IsOpen, &group.NewMembersCanCreateEvents,
			&group.EventApprovalMode, &group.EventAttendeeManagementMode, &group.ContributorEventsRequired,
			&group.CreatedAt, &group.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning group: %w", err)
		}
		groups = append(groups, &group)
	}

	return groups, nil
}

// UpdateGroup updates a group
func (d *Database) UpdateGroup(group *models.Group) error {
	// Update the group
	_, err := d.db.Exec(`
		UPDATE groups
		SET
			name = ?,
			description = ?,
			channel_id = ?,
			is_open = ?,
			new_members_can_create_events = ?,
			event_approval_mode = ?,
			event_attendee_management_mode = ?,
			contributor_events_required = ?,
			updated_at = CURRENT_TIMESTAMP
		WHERE group_id = ?
	`,
		group.Name, group.Description, group.ChannelID, group.IsOpen, group.NewMembersCanCreateEvents,
		group.EventApprovalMode, group.EventAttendeeManagementMode, group.ContributorEventsRequired,
		group.GroupID,
	)
	if err != nil {
		return fmt.Errorf("error updating group: %w", err)
	}

	return nil
}

// DeleteGroup deletes a group
func (d *Database) DeleteGroup(groupID int64) error {
	// Delete the group
	_, err := d.db.Exec(`
		DELETE FROM groups
		WHERE group_id = ?
	`, groupID)
	if err != nil {
		return fmt.Errorf("error deleting group: %w", err)
	}

	return nil
}

// Group member methods

// AddGroupMember adds a member to a group
func (d *Database) AddGroupMember(groupID int64, userID string, isLeader bool) error {
	// Insert the member
	_, err := d.db.Exec(`
		INSERT INTO group_members (
			group_id, user_id, is_leader
		) VALUES (?, ?, ?)
	`,
		groupID, userID, isLeader,
	)
	if err != nil {
		return fmt.Errorf("error adding group member: %w", err)
	}

	return nil
}

// GetGroupMember gets a member of a group
func (d *Database) GetGroupMember(groupID int64, userID string) (*models.GroupMember, error) {
	// Query the member
	row := d.db.QueryRow(`
		SELECT
			group_id, user_id, is_leader, is_contributor, joined_at, updated_at
		FROM group_members
		WHERE group_id = ? AND user_id = ?
	`, groupID, userID)

	// Scan the result
	var member models.GroupMember
	err := row.Scan(
		&member.GroupID, &member.UserID, &member.IsLeader, &member.IsContributor, &member.JoinedAt, &member.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("error getting group member: %w", err)
	}

	return &member, nil
}

// GetGroupMembers gets all members of a group
func (d *Database) GetGroupMembers(groupID int64) ([]*models.GroupMember, error) {
	// Query the members
	rows, err := d.db.Query(`
		SELECT
			group_id, user_id, is_leader, is_contributor, joined_at, updated_at
		FROM group_members
		WHERE group_id = ?
		ORDER BY is_leader DESC, joined_at
	`, groupID)
	if err != nil {
		return nil, fmt.Errorf("error getting group members: %w", err)
	}
	defer rows.Close()

	// Scan the results
	var members []*models.GroupMember
	for rows.Next() {
		var member models.GroupMember
		err := rows.Scan(
			&member.GroupID, &member.UserID, &member.IsLeader, &member.IsContributor, &member.JoinedAt, &member.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning group member: %w", err)
		}
		members = append(members, &member)
	}

	return members, nil
}

// UpdateGroupMember updates a member of a group
func (d *Database) UpdateGroupMember(member *models.GroupMember) error {
	// Update the member
	_, err := d.db.Exec(`
		UPDATE group_members
		SET
			is_leader = ?,
			is_contributor = ?,
			updated_at = CURRENT_TIMESTAMP
		WHERE group_id = ? AND user_id = ?
	`,
		member.IsLeader, member.IsContributor, member.GroupID, member.UserID,
	)
	if err != nil {
		return fmt.Errorf("error updating group member: %w", err)
	}

	return nil
}

// RemoveGroupMember removes a member from a group
func (d *Database) RemoveGroupMember(groupID int64, userID string) error {
	// Delete the member
	_, err := d.db.Exec(`
		DELETE FROM group_members
		WHERE group_id = ? AND user_id = ?
	`, groupID, userID)
	if err != nil {
		return fmt.Errorf("error removing group member: %w", err)
	}

	return nil
}

// Event methods

// CreateEvent creates a new event
func (d *Database) CreateEvent(event *models.Event) (int64, error) {
	// Insert the event
	result, err := d.db.Exec(`
		INSERT INTO events (
			group_id, host_id, name, description, date_time, location_name, location_address,
			max_attendees, status, message_id, thread_id
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		event.GroupID, event.HostID, event.Name, event.Description, event.DateTime, event.LocationName, event.LocationAddress,
		event.MaxAttendees, event.Status, event.MessageID, event.ThreadID,
	)
	if err != nil {
		return 0, fmt.Errorf("error creating event: %w", err)
	}

	// Get the event ID
	eventID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("error getting event ID: %w", err)
	}

	return eventID, nil
}

// GetEvent gets an event by ID
func (d *Database) GetEvent(eventID int64) (*models.Event, error) {
	// Query the event
	row := d.db.QueryRow(`
		SELECT
			event_id, group_id, host_id, name, description, date_time, location_name, location_address,
			max_attendees, status, message_id, thread_id, created_at, updated_at
		FROM events
		WHERE event_id = ?
	`, eventID)

	// Scan the result
	var event models.Event
	err := row.Scan(
		&event.EventID, &event.GroupID, &event.HostID, &event.Name, &event.Description, &event.DateTime, &event.LocationName, &event.LocationAddress,
		&event.MaxAttendees, &event.Status, &event.MessageID, &event.ThreadID, &event.CreatedAt, &event.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("error getting event: %w", err)
	}

	return &event, nil
}

// GetEvents gets all events for a group
func (d *Database) GetEvents(groupID int64) ([]*models.Event, error) {
	// Query the events
	rows, err := d.db.Query(`
		SELECT
			event_id, group_id, host_id, name, description, date_time, location_name, location_address,
			max_attendees, status, message_id, thread_id, created_at, updated_at
		FROM events
		WHERE group_id = ?
		ORDER BY date_time
	`, groupID)
	if err != nil {
		return nil, fmt.Errorf("error getting events: %w", err)
	}
	defer rows.Close()

	// Scan the results
	var events []*models.Event
	for rows.Next() {
		var event models.Event
		err := rows.Scan(
			&event.EventID, &event.GroupID, &event.HostID, &event.Name, &event.Description, &event.DateTime, &event.LocationName, &event.LocationAddress,
			&event.MaxAttendees, &event.Status, &event.MessageID, &event.ThreadID, &event.CreatedAt, &event.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning event: %w", err)
		}
		events = append(events, &event)
	}

	return events, nil
}

// GetUpcomingEvents gets upcoming events for a group
func (d *Database) GetUpcomingEvents(groupID int64) ([]*models.Event, error) {
	// Query the events
	rows, err := d.db.Query(`
		SELECT
			event_id, group_id, host_id, name, description, date_time, location_name, location_address,
			max_attendees, status, message_id, thread_id, created_at, updated_at
		FROM events
		WHERE group_id = ? AND date_time > CURRENT_TIMESTAMP AND status = ?
		ORDER BY date_time
	`, groupID, string(models.EventStatusApproved))
	if err != nil {
		return nil, fmt.Errorf("error getting upcoming events: %w", err)
	}
	defer rows.Close()

	// Scan the results
	var events []*models.Event
	for rows.Next() {
		var event models.Event
		err := rows.Scan(
			&event.EventID, &event.GroupID, &event.HostID, &event.Name, &event.Description, &event.DateTime, &event.LocationName, &event.LocationAddress,
			&event.MaxAttendees, &event.Status, &event.MessageID, &event.ThreadID, &event.CreatedAt, &event.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning event: %w", err)
		}
		events = append(events, &event)
	}

	return events, nil
}

// UpdateEvent updates an event
func (d *Database) UpdateEvent(event *models.Event) error {
	// Update the event
	_, err := d.db.Exec(`
		UPDATE events
		SET
			group_id = ?,
			host_id = ?,
			name = ?,
			description = ?,
			date_time = ?,
			location_name = ?,
			location_address = ?,
			max_attendees = ?,
			status = ?,
			message_id = ?,
			thread_id = ?,
			updated_at = CURRENT_TIMESTAMP
		WHERE event_id = ?
	`,
		event.GroupID, event.HostID, event.Name, event.Description, event.DateTime, event.LocationName, event.LocationAddress,
		event.MaxAttendees, event.Status, event.MessageID, event.ThreadID, event.EventID,
	)
	if err != nil {
		return fmt.Errorf("error updating event: %w", err)
	}

	return nil
}

// DeleteEvent deletes an event
func (d *Database) DeleteEvent(eventID int64) error {
	// Delete the event
	_, err := d.db.Exec(`
		DELETE FROM events
		WHERE event_id = ?
	`, eventID)
	if err != nil {
		return fmt.Errorf("error deleting event: %w", err)
	}

	return nil
}

// Event attendee methods

// AddEventAttendee adds an attendee to an event
func (d *Database) AddEventAttendee(eventID int64, userID string, rsvpStatus string) error {
	// Insert the attendee
	_, err := d.db.Exec(`
		INSERT INTO event_attendees (
			event_id, user_id, rsvp_status
		) VALUES (?, ?, ?)
	`,
		eventID, userID, rsvpStatus,
	)
	if err != nil {
		return fmt.Errorf("error adding event attendee: %w", err)
	}

	return nil
}

// GetEventAttendee gets an attendee of an event
func (d *Database) GetEventAttendee(eventID int64, userID string) (*models.EventAttendee, error) {
	// Query the attendee
	row := d.db.QueryRow(`
		SELECT
			event_id, user_id, rsvp_status, rsvp_time, updated_at
		FROM event_attendees
		WHERE event_id = ? AND user_id = ?
	`, eventID, userID)

	// Scan the result
	var attendee models.EventAttendee
	err := row.Scan(
		&attendee.EventID, &attendee.UserID, &attendee.RSVPStatus, &attendee.RSVPTime, &attendee.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("error getting event attendee: %w", err)
	}

	return &attendee, nil
}

// GetEventAttendees gets all attendees of an event
func (d *Database) GetEventAttendees(eventID int64) ([]*models.EventAttendee, error) {
	// Query the attendees
	rows, err := d.db.Query(`
		SELECT
			event_id, user_id, rsvp_status, rsvp_time, updated_at
		FROM event_attendees
		WHERE event_id = ?
		ORDER BY rsvp_time
	`, eventID)
	if err != nil {
		return nil, fmt.Errorf("error getting event attendees: %w", err)
	}
	defer rows.Close()

	// Scan the results
	var attendees []*models.EventAttendee
	for rows.Next() {
		var attendee models.EventAttendee
		err := rows.Scan(
			&attendee.EventID, &attendee.UserID, &attendee.RSVPStatus, &attendee.RSVPTime, &attendee.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning event attendee: %w", err)
		}
		attendees = append(attendees, &attendee)
	}

	return attendees, nil
}

// UpdateEventAttendee updates an attendee of an event
func (d *Database) UpdateEventAttendee(attendee *models.EventAttendee) error {
	// Update the attendee
	_, err := d.db.Exec(`
		UPDATE event_attendees
		SET
			rsvp_status = ?,
			updated_at = CURRENT_TIMESTAMP
		WHERE event_id = ? AND user_id = ?
	`,
		attendee.RSVPStatus, attendee.EventID, attendee.UserID,
	)
	if err != nil {
		return fmt.Errorf("error updating event attendee: %w", err)
	}

	return nil
}

// RemoveEventAttendee removes an attendee from an event
func (d *Database) RemoveEventAttendee(eventID int64, userID string) error {
	// Delete the attendee
	_, err := d.db.Exec(`
		DELETE FROM event_attendees
		WHERE event_id = ? AND user_id = ?
	`, eventID, userID)
	if err != nil {
		return fmt.Errorf("error removing event attendee: %w", err)
	}

	return nil
}

// Settings methods

// GetSettings gets the settings for a guild
func (d *Database) GetSettings(guildID string) (*models.Settings, error) {
	// Query the settings
	row := d.db.QueryRow(`
		SELECT
			guild_id, terminology, updated_at
		FROM settings
		WHERE guild_id = ?
	`, guildID)

	// Scan the result
	var settings models.Settings
	var terminologyJSON string
	err := row.Scan(
		&settings.GuildID, &terminologyJSON, &settings.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("error getting settings: %w", err)
	}

	// Parse the terminology JSON
	if terminologyJSON != "" {
		err = json.Unmarshal([]byte(terminologyJSON), &settings.Terminology)
		if err != nil {
			log.Printf("Error parsing terminology JSON: %v", err)
		}
	}

	return &settings, nil
}

// UpdateSettings updates the settings for a guild
func (d *Database) UpdateSettings(settings *models.Settings) error {
	// Marshal the terminology to JSON
	terminologyJSON, err := json.Marshal(settings.Terminology)
	if err != nil {
		return fmt.Errorf("error marshaling terminology: %w", err)
	}

	// Check if settings exist
	var count int
	err = d.db.QueryRow(`
		SELECT COUNT(*)
		FROM settings
		WHERE guild_id = ?
	`, settings.GuildID).Scan(&count)
	if err != nil {
		return fmt.Errorf("error checking if settings exist: %w", err)
	}

	if count == 0 {
		// Insert the settings
		_, err = d.db.Exec(`
			INSERT INTO settings (
				guild_id, terminology
			) VALUES (?, ?)
		`,
			settings.GuildID, string(terminologyJSON),
		)
		if err != nil {
			return fmt.Errorf("error inserting settings: %w", err)
		}
	} else {
		// Update the settings
		_, err = d.db.Exec(`
			UPDATE settings
			SET
				terminology = ?,
				updated_at = CURRENT_TIMESTAMP
			WHERE guild_id = ?
		`,
			string(terminologyJSON), settings.GuildID,
		)
		if err != nil {
			return fmt.Errorf("error updating settings: %w", err)
		}
	}

	return nil
} 