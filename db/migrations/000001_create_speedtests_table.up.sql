CREATE TABLE speedtests (
  id INTEGER
    PRIMARY KEY ASC AUTOINCREMENT
    ,
  started_at TEXT
    NOT NULL
    CHECK (DATETIME(started_at) IS NOT NULL) -- Ensure it's a valid datetime
    ,
  ended_at TEXT
    NOT NULL
    CHECK (DATETIME(started_at) IS NOT NULL) -- Ensure it's a valid datetime
    ,
  kbps_down INTEGER
    NOT NULL
    CHECK(kbps_down > 0)
);
