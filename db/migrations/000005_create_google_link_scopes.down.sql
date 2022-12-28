DELETE TABLE google_link_scopes;

ALTER TABLE
  google_links
ADD
  scope TEXT NOT NULL;
