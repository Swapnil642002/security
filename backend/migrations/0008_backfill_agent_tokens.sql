UPDATE employee_laptops
SET agent_token = md5(id::text || ':' || hostname || ':' || created_at::text)
WHERE agent_token IS NULL OR btrim(agent_token) = '';
