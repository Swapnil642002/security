CREATE TABLE IF NOT EXISTS enrollment_links (
    id BIGSERIAL PRIMARY KEY,
    token TEXT NOT NULL UNIQUE,
    created_by BIGINT NOT NULL REFERENCES users(id),
    expires_at TIMESTAMPTZ NOT NULL,
    max_uses INT NOT NULL DEFAULT 1,
    used_count INT NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    require_approval BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS device_enrollments (
    id BIGSERIAL PRIMARY KEY,
    link_id BIGINT NOT NULL REFERENCES enrollment_links(id),
    status TEXT NOT NULL CHECK (status IN ('pending', 'approved', 'disabled', 'rejected')) DEFAULT 'pending',
    hostname TEXT NOT NULL,
    employee_name TEXT NOT NULL,
    employee_email TEXT NOT NULL,
    os_type TEXT NOT NULL CHECK (os_type IN ('windows', 'macos', 'linux')),
    current_ip TEXT,
    fingerprint TEXT,
    laptop_id BIGINT REFERENCES employee_laptops(id),
    approved_by BIGINT REFERENCES users(id),
    approved_at TIMESTAMPTZ,
    disabled_by BIGINT REFERENCES users(id),
    disabled_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_device_enrollments_status ON device_enrollments(status);
CREATE INDEX IF NOT EXISTS idx_device_enrollments_employee_email ON device_enrollments(employee_email);

CREATE TABLE IF NOT EXISTS system_notifications (
    id BIGSERIAL PRIMARY KEY,
    type TEXT NOT NULL,
    message TEXT NOT NULL,
    target_role TEXT NOT NULL DEFAULT 'admin',
    is_read BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_system_notifications_created_at ON system_notifications(created_at DESC);
