CREATE TABLE google_link_scopes (
  id INTEGER PRIMARY KEY ASC AUTOINCREMENT,
  google_link_id REFERENCES google_links(id) ON DELETE CASCADE,
  scope TEXT NOT NULL
);

ALTER TABLE
  google_links DROP scope;
