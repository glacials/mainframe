ALTER TABLE google_links
DROP COLUMN expires_at;

ALTER TABLE google_links
ADD COLUMN expires_at TEXT NOT NULL CHECK (DATETIME(expires_at) IS NOT NULL);
