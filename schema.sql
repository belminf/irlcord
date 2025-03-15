-- 1. Users Table
CREATE TABLE IF NOT EXISTS Users (
    user_id VARCHAR(255) PRIMARY KEY,
    venmo_username VARCHAR(255),
    dietary_restrictions TEXT,
    email VARCHAR(255),
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 2. Groups Table
CREATE TABLE IF NOT EXISTS Groups (
    group_id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(255) UNIQUE,
    description TEXT,
    is_open BOOLEAN DEFAULT TRUE,
    chat_inactivity_days INTEGER DEFAULT 0,
    event_inactivity_days INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    channel_id VARCHAR(255),
    contributor_events_required INTEGER DEFAULT 3,
    new_member_deposit DECIMAL(10,2),
    new_members_can_create_events BOOLEAN DEFAULT TRUE,
    event_approval_mode VARCHAR(10) DEFAULT 'public',
    event_attendee_management_mode VARCHAR(10) DEFAULT 'host'
);

-- 3. Group Members Table
CREATE TABLE IF NOT EXISTS GroupMembers (
    group_id INTEGER,
    user_id VARCHAR(255),
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_leader BOOLEAN DEFAULT FALSE,
    FOREIGN KEY (group_id) REFERENCES Groups(group_id),
    FOREIGN KEY (user_id) REFERENCES Users(user_id),
    PRIMARY KEY (group_id, user_id)
);

-- 4. Events Table
CREATE TABLE IF NOT EXISTS Events (
    event_id INTEGER PRIMARY KEY AUTOINCREMENT,
    group_id INTEGER,
    host_id VARCHAR(255),
    name VARCHAR(255),
    date_time TIMESTAMP,
    location_name VARCHAR(255),
    location_address TEXT,
    description TEXT,
    max_attendees INTEGER,
    is_public BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    thread_id VARCHAR(255),
    status VARCHAR(10) DEFAULT 'pending',
    FOREIGN KEY (group_id) REFERENCES Groups(group_id),
    FOREIGN KEY (host_id) REFERENCES Users(user_id)
);

-- 5. Event Attendees Table
CREATE TABLE IF NOT EXISTS EventAttendees (
    event_id INTEGER,
    user_id VARCHAR(255),
    rsvp_status VARCHAR(10) DEFAULT 'ATTENDING',
    calendar_event_id VARCHAR(255),
    plus_one_calendar_event_id VARCHAR(255),
    FOREIGN KEY (event_id) REFERENCES Events(event_id),
    FOREIGN KEY (user_id) REFERENCES Users(user_id),
    PRIMARY KEY (event_id, user_id)
);

-- 6. Bills Table
CREATE TABLE IF NOT EXISTS Bills (
    bill_id INTEGER PRIMARY KEY AUTOINCREMENT,
    event_id INTEGER,
    user_id VARCHAR(255),
    amount DECIMAL(10,2),
    paid BOOLEAN DEFAULT FALSE,
    FOREIGN KEY (event_id) REFERENCES Events(event_id),
    FOREIGN KEY (user_id) REFERENCES Users(user_id)
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_group_members_user_id ON GroupMembers(user_id);
CREATE INDEX IF NOT EXISTS idx_events_group_id ON Events(group_id);
CREATE INDEX IF NOT EXISTS idx_events_host_id ON Events(host_id);
CREATE INDEX IF NOT EXISTS idx_event_attendees_user_id ON EventAttendees(user_id);
CREATE INDEX IF NOT EXISTS idx_bills_event_id ON Bills(event_id);
CREATE INDEX IF NOT EXISTS idx_bills_user_id ON Bills(user_id); 