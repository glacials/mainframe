ALTER TABLE
  google_links
ADD
  COLUMN google_user_id REFERENCES google_users(id) ON DELETE CASCADE;
