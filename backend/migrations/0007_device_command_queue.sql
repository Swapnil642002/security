ALTER TABLE employee_laptops
    ADD COLUMN IF NOT EXISTS agent_token TEXT UNIQUE,
    ADD COLUMN IF NOT EXISTS usb_storage_blocked BOOLEAN NOT NULL DEFAULT FALSE;

CREATE TABLE IF NOT EXISTS device_commands (
    id BIGSERIAL PRIMARY KEY,
    laptop_id BIGINT NOT NULL REFERENCES employee_laptops(id) ON DELETE CASCADE,
    command_type TEXT NOT NULL CHECK (command_type IN ('usb.block', 'usb.unblock')),
    payload_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    status TEXT NOT NULL CHECK (status IN ('pending', 'in_progress', 'success', 'failed')) DEFAULT 'pending',
    result_text TEXT NOT NULL DEFAULT '',
    created_by BIGINT NOT NULL REFERENCES users(id),
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_device_commands_laptop_status ON device_commands(laptop_id, status, created_at);
CREATE INDEX IF NOT EXISTS idx_device_commands_status_created_at ON device_commands(status, created_at);
