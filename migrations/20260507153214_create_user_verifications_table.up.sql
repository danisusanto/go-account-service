CREATE TYPE user_verification_type AS ENUM (
  'EMAIL_VERIFICATION',
  'RESET_PASSWORD',
  'TWO_FACTOR_AUTH',
  'CHANGE_EMAIL',
  'CHANGE_PHONE'
);

CREATE TABLE user_verifications (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  verification_type user_verification_type DEFAULT 'EMAIL_VERIFICATION' NOT NULL,
  token_hash VARCHAR NOT NULL UNIQUE,
  attempt INT DEFAULT 0,
  payload TEXT,
  created_by BIGINT,
  updated_by BIGINT,
  deleted_by BIGINT,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL,
  verified_at TIMESTAMP,
  used_at TIMESTAMP,
  expired_at TIMESTAMP NOT NULL,
  deleted_at TIMESTAMP
)
