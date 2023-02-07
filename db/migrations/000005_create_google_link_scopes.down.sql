DROP TABLE google_link_scopes;

ALTER TABLE
  google_links
ADD COLUMN
  scope TEXT NOT NULL DEFAULT '';
