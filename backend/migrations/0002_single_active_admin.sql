CREATE UNIQUE INDEX IF NOT EXISTS uniq_single_active_admin
ON users (role)
WHERE role = 'admin' AND is_active = TRUE;
