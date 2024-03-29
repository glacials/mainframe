CREATE TABLE google_users (
  id INTEGER PRIMARY KEY ASC AUTOINCREMENT,
  user_id REFERENCES users(id) ON DELETE CASCADE,
  external_id TEXT NOT NULL UNIQUE,
  email TEXT NOT NULL,
  verified_email INTEGER NOT NULL,
  family_name TEXT NOT NULL,
  given_name TEXT NOT NULL,
  name string NOT NULL,
  picture TEXT NOT NULL,
  gender TEXT NOT NULL,
  hosted_domain TEXT NOT NULL,
  link TEXT NOT NULL,
  locale TEXT NOT NULL,
  created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP CHECK (DATETIME(created_at) IS NOT NULL),
  updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP CHECK (DATETIME(updated_at) IS NOT NULL)
);
