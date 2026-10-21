-- Performance indexes for scaling to 50k users
-- idx_todos_user_id: speeds up the primary query (all tasks for a user)
CREATE INDEX IF NOT EXISTS idx_todos_user_id ON todos (user_id);

-- idx_todos_user_status: supports filtered queries by status
CREATE INDEX IF NOT EXISTS idx_todos_user_id_status ON todos (user_id, status);

-- idx_todos_user_created: supports ORDER BY created_at DESC with pagination
CREATE INDEX IF NOT EXISTS idx_todos_user_id_created ON todos (user_id, created_at DESC);

-- idx_users_email: already covered by UNIQUE constraint, but explicit for clarity
-- (no action needed – the UNIQUE constraint creates an implicit B-tree index)
