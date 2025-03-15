-- 1. Users Table
CREATE TABLE Users (
    user_id VARCHAR(255) PRIMARY KEY,
    venmo_username VARCHAR(255),
    dietary_restrictions TEXT,
    email VARCHAR(255),
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 2. Groups Table
CREATE TABLE Groups (
    group_id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) UNIQUE,
    description TEXT,
    is_open BOOLEAN DEFAULT TRUE,
    chat_inactivity_days INT DEFAULT 0,
    event_inactivity_days INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    channel_id VARCHAR(255),
    contributor_events_required INT DEFAULT 3,
    new_member_deposit DECIMAL(10,2),
    new_members_can_create_events BOOLEAN DEFAULT TRUE,
    event_approval_mode ENUM('none', 'public', 'all') DEFAULT 'public',
    event_attendee_management_mode ENUM('host', 'self') DEFAULT 'host'
);

-- 3. Group Members Table
CREATE TABLE GroupMembers (
    group_id INT,
    user_id VARCHAR(255),
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_leader BOOLEAN DEFAULT FALSE,
    FOREIGN KEY (group_id) REFERENCES Groups(group_id),
    FOREIGN KEY (user_id) REFERENCES Users(user_id),
    PRIMARY KEY (group_id, user_id)
);

-- 4. Events Table
CREATE TABLE Events (
    event_id INT AUTO_INCREMENT PRIMARY KEY,
    group_id INT,
    host_id VARCHAR(255),
    name VARCHAR(255),
    date_time TIMESTAMP,
    location_name VARCHAR(255),
    location_address TEXT,
    description TEXT,
    max_attendees INT,
    is_public BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    thread_id VARCHAR(255),
    status ENUM('pending', 'approved', 'rejected') DEFAULT 'pending',
    FOREIGN KEY (group_id) REFERENCES Groups(group_id),
    FOREIGN KEY (host_id) REFERENCES Users(user_id)
);

-- 5. Event Attendees Table
CREATE TABLE EventAttendees (
    event_id INT,
    user_id VARCHAR(255),
    rsvp_status ENUM('ATTENDING', 'WAITLIST', 'DECLINED') DEFAULT 'ATTENDING',
    calendar_event_id VARCHAR(255),
    plus_one_calendar_event_id VARCHAR(255),
    FOREIGN KEY (event_id) REFERENCES Events(event_id),
    FOREIGN KEY (user_id) REFERENCES Users(user_id),
    PRIMARY KEY (event_id, user_id)
);

-- 6. Bills Table
CREATE TABLE Bills (
    bill_id INT AUTO_INCREMENT PRIMARY KEY,
    event_id INT,
    user_id VARCHAR(255),
    amount DECIMAL(10,2),
    paid BOOLEAN DEFAULT FALSE,
    FOREIGN KEY (event_id) REFERENCES Events(event_id),
    FOREIGN KEY (user_id) REFERENCES Users(user_id)
);
