CREATE TABLE google_links (
  id INTEGER
    PRIMARY KEY ASC AUTOINCREMENT
    ,
  access_token TEXT
    NOT NULL
    ,
  token_type TEXT
    NOT NULL
    ,
  refresh_token TEXT
    NOT NULL
    ,
  expires_at TEXT
    NOT NULL,
    CHECK (DATETIME(expires_at) IS NOT NULL) -- Ensure it's a valid datetime
    ,
  scope TEXT
    NOT NULL
    ,
  created_at TEXT
    NOT NULL DEFAULT NOW()
    CHECK (DATETIME(created_at) IS NOT NULL) -- Ensure it's a valid datetime
    ,
  updated_at TEXT
    NOT NULL DEFAULT NOW()
    CHECK (DATETIME(updated_at) IS NOT NULL) -- Ensure it's a valid datetime
    ,
);
