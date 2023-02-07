CREATE TABLE ip_addresses (
  id INTEGER
    PRIMARY KEY ASC AUTOINCREMENT
    ,
  ip_address TEXT
    NOT NULL
    ,
  created_at TEXT
    NOT NULL
    DEFAULT CURRENT_TIMESTAMP
    CHECK (DATETIME(created_at) IS NOT NULL)
);
