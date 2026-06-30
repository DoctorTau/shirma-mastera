CREATE TABLE users (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    email text NOT NULL UNIQUE,
    password_hash text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

-- Pre-auth dev/test encounters have no owner; drop them before enforcing NOT NULL.
DELETE FROM encounters;

ALTER TABLE encounters ADD COLUMN user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE;
CREATE INDEX idx_encounters_user_id ON encounters (user_id);
