CREATE TABLE users (
  id BIGSERIAL PRIMARY KEY,
  
  name text NOT NULL UNIQUE,
  email text NOT NULL,
  
  password_hash text,
  
  created_at timestamptz NOT NULL DEFAULT now(),
  
  is_active boolean NOT NULL DEFAULT false
);

CREATE UNIQUE INDEX users_email_lower_idx ON users (lower(email));
CREATE INDEX users_is_active_idx ON users (is_active);