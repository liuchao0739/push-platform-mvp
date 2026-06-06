-- Push Platform MVP 数据库初始化

CREATE TABLE IF NOT EXISTS devices (
    id BIGSERIAL PRIMARY KEY,
    device_token VARCHAR(255) NOT NULL UNIQUE,
    user_id VARCHAR(128),
    platform VARCHAR(32) NOT NULL DEFAULT 'unknown',  -- ios / android / harmony / web
    app_id VARCHAR(128) NOT NULL,
    status VARCHAR(16) NOT NULL DEFAULT 'active',     -- active / inactive
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_devices_token ON devices(device_token);
CREATE INDEX idx_devices_user_id ON devices(user_id);
CREATE INDEX idx_devices_app_id ON devices(app_id);

CREATE TABLE IF NOT EXISTS push_messages (
    id BIGSERIAL PRIMARY KEY,
    msg_id VARCHAR(64) NOT NULL UNIQUE,
    app_id VARCHAR(128) NOT NULL,
    target_user_id VARCHAR(128),
    target_device_token VARCHAR(255),
    title VARCHAR(256),
    body TEXT,
    status VARCHAR(16) NOT NULL DEFAULT 'pending',  -- pending / sent / delivered / failed
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    delivered_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_push_messages_msg_id ON push_messages(msg_id);
CREATE INDEX idx_push_messages_status ON push_messages(status);
CREATE INDEX idx_push_messages_target ON push_messages(target_user_id, target_device_token);
