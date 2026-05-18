CREATE TABLE IF NOT EXISTS departments (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS employee_laptops (
    id BIGSERIAL PRIMARY KEY,
    hostname TEXT NOT NULL UNIQUE,
    employee_name TEXT NOT NULL,
    employee_email TEXT NOT NULL,
    os_type TEXT NOT NULL CHECK (os_type IN ('windows', 'macos', 'linux')),
    department_id BIGINT REFERENCES departments(id),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    last_seen_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_employee_laptops_department_id ON employee_laptops(department_id);
CREATE INDEX IF NOT EXISTS idx_employee_laptops_employee_email ON employee_laptops(employee_email);

CREATE TABLE IF NOT EXISTS policy_assignments (
    id BIGSERIAL PRIMARY KEY,
    policy_id BIGINT NOT NULL REFERENCES firewall_policies(id) ON DELETE CASCADE,
    assignment_type TEXT NOT NULL CHECK (assignment_type IN ('department', 'laptop')),
    department_id BIGINT REFERENCES departments(id) ON DELETE CASCADE,
    laptop_id BIGINT REFERENCES employee_laptops(id) ON DELETE CASCADE,
    is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (
        (assignment_type = 'department' AND department_id IS NOT NULL AND laptop_id IS NULL) OR
        (assignment_type = 'laptop' AND laptop_id IS NOT NULL AND department_id IS NULL)
    )
);

CREATE INDEX IF NOT EXISTS idx_policy_assignments_policy_id ON policy_assignments(policy_id);
CREATE INDEX IF NOT EXISTS idx_policy_assignments_department_id ON policy_assignments(department_id);
CREATE INDEX IF NOT EXISTS idx_policy_assignments_laptop_id ON policy_assignments(laptop_id);
