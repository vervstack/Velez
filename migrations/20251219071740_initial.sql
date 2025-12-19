-- +goose Up
-- Create the schema
CREATE SCHEMA IF NOT EXISTS velez;

-- Create the user (replace 'your_secure_password' with an actual password)
-- Check if user exists first or use a DO block in PostgreSQL
DO $$
    BEGIN
        IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'velez') THEN
            CREATE ROLE velez WITH LOGIN PASSWORD 'your_secure_password';
        END IF;
    END
$$;

-- Grant usage on the schema
GRANT USAGE ON SCHEMA velez TO velez;

-- Grant full CRUD and table management permissions
GRANT ALL PRIVILEGES ON SCHEMA velez TO velez;

-- Ensure future tables created in this schema are also accessible by the user
ALTER DEFAULT PRIVILEGES IN SCHEMA velez GRANT ALL ON TABLES TO velez;
ALTER DEFAULT PRIVILEGES IN SCHEMA velez GRANT ALL ON SEQUENCES TO velez;

-- +goose Down
-- Reverting the changes: remove privileges and schema
-- Note: Dropping the role might fail if other objects depend on it
ALTER DEFAULT PRIVILEGES IN SCHEMA velez REVOKE ALL ON TABLES FROM velez;
ALTER DEFAULT PRIVILEGES IN SCHEMA velez REVOKE ALL ON SEQUENCES FROM velez;
DROP SCHEMA IF EXISTS velez CASCADE;
-- DROP ROLE velez; -- Uncomment if you want the user deleted on rollback