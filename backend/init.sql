-- Event Manager PostgreSQL Schema
-- Uses JSONB for flexible configuration storage (similar to Firestore)

-- Events table
CREATE TABLE IF NOT EXISTS events (
    event_id        VARCHAR(36) PRIMARY KEY,
    type            VARCHAR(20) NOT NULL CHECK (type IN ('VOTE', 'LINEUP', 'MEMO')),
    title           VARCHAR(255) NOT NULL,
    is_active       BOOLEAN DEFAULT true,
    is_archived     BOOLEAN DEFAULT false,
    created_by      VARCHAR(50) NOT NULL,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    config          JSONB NOT NULL DEFAULT '{}'
);

CREATE INDEX IF NOT EXISTS idx_events_created_at ON events(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_events_type ON events(type);
CREATE INDEX IF NOT EXISTS idx_events_is_active ON events(is_active);
CREATE INDEX IF NOT EXISTS idx_events_is_archived ON events(is_archived);


-- Interactions table (like Firestore subcollection)
CREATE TABLE IF NOT EXISTS interactions (
    id                  VARCHAR(100) PRIMARY KEY DEFAULT gen_random_uuid()::text,
    event_id            VARCHAR(36) NOT NULL REFERENCES events(event_id) ON DELETE CASCADE,
    user_id             VARCHAR(50) NOT NULL,
    type                VARCHAR(20) NOT NULL CHECK (type IN ('VOTE', 'LINEUP', 'MEMO')),
    user_display_name   VARCHAR(100),
    user_picture_url    TEXT,
    status              VARCHAR(20),
    timestamp           TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    -- JSONB payload for flexible fields
    payload             JSONB NOT NULL DEFAULT '{}'
);

CREATE INDEX IF NOT EXISTS idx_interactions_event ON interactions(event_id);
CREATE INDEX IF NOT EXISTS idx_interactions_user ON interactions(user_id);
CREATE INDEX IF NOT EXISTS idx_interactions_type ON interactions(type);
CREATE INDEX IF NOT EXISTS idx_interactions_event_type ON interactions(event_id, type);
CREATE INDEX IF NOT EXISTS idx_interactions_event_user_type ON interactions(event_id, user_id, type);

-- Users table
CREATE TABLE IF NOT EXISTS users (
    line_user_id        VARCHAR(50) PRIMARY KEY,
    line_display_name   VARCHAR(100),
    picture_url         TEXT,
    role                VARCHAR(20) DEFAULT 'user' CHECK (role IN ('user', 'admin')),
    created_at          TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Comments to document JSONB field structure
COMMENT ON COLUMN events.config IS 'JSON structure: {
    "allowMultiSelect": boolean,
    "maxVotes": int,
    "options": string[],
    "maxParticipants": int,
    "waitlistLimit": int,
    "maxCountPerUser": int,
    "startTime": timestamp,
    "endTime": timestamp,
    "maxCommentsPerUser": int,
    "allowReaction": boolean
}';

COMMENT ON COLUMN interactions.payload IS 'JSON structure varies by type:
VOTE: {"selectedOptions": string[]}
LINEUP: {"count": int, "note": string, "cancelledAt": timestamp}
MEMO: {"content": string, "clapCount": int, "reactions": string[]}
';
