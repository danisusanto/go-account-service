CREATE TYPE role_type AS ENUM (
	'ADMIN',
  'MANAGER',
  'STAFF',
  'VIEWER'
);

CREATE TABLE users (
  id BIGSERIAL PRIMARY KEY,
  name VARCHAR NOT NULL,
  email VARCHAR UNIQUE NOT NULL,
  password VARCHAR NOT NULL,
  country_code VARCHAR(3) NOT NULL,
  phone_number VARCHAR(100) NOT NULL,
  google_id VARCHAR(100),
  is_active BOOLEAN DEFAULT TRUE,
  role role_type DEFAULT 'ADMIN' NOT NULL,
  created_by BIGINT,
  updated_by BIGINT,
  deleted_by BIGINT,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL,
  email_verified_at TIMESTAMP,
  deleted_at TIMESTAMP
);
