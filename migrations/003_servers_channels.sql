-- Servers table
CREATE TABLE IF NOT EXISTS servers (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    owner_id INTEGER NOT NULL REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Server members (who can access which servers)
CREATE TABLE IF NOT EXISTS server_members (
    id SERIAL PRIMARY KEY,
    server_id INTEGER NOT NULL REFERENCES servers(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(server_id, user_id)
);

-- Channels table
CREATE TABLE IF NOT EXISTS channels (
    id SERIAL PRIMARY KEY,
    server_id INTEGER REFERENCES servers(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(20) DEFAULT 'text', -- 'text' or 'dm'
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- DM participants (for direct messages)
CREATE TABLE IF NOT EXISTS dm_participants (
    id SERIAL PRIMARY KEY,
    channel_id INTEGER NOT NULL REFERENCES channels(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE(channel_id, user_id)
);

-- Update messages table to reference channels properly
ALTER TABLE messages DROP CONSTRAINT IF EXISTS messages_channel_id_fkey;
ALTER TABLE messages ADD CONSTRAINT messages_channel_id_fkey 
    FOREIGN KEY (channel_id) REFERENCES channels(id) ON DELETE CASCADE;

-- Create default server and channel
INSERT INTO servers (id, name, owner_id) VALUES (1, 'Nexus Server', 1) ON CONFLICT DO NOTHING;
INSERT INTO channels (id, server_id, name, type) VALUES (1, 1, 'general', 'text') ON CONFLICT DO NOTHING;

-- Indexes
CREATE INDEX idx_server_members_server ON server_members(server_id);
CREATE INDEX idx_server_members_user ON server_members(user_id);
CREATE INDEX idx_channels_server ON channels(server_id);
CREATE INDEX idx_dm_participants_channel ON dm_participants(channel_id);
CREATE INDEX idx_dm_participants_user ON dm_participants(user_id);