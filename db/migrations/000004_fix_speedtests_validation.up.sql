-- Previous speedtests table def CHECKed ended_at using DATETIME(started_at)
CREATE TABLE speedtests_tmp (
  id INTEGER
    PRIMARY KEY ASC AUTOINCREMENT
    ,
  started_at TEXT
    NOT NULL
    CHECK (DATETIME(started_at) IS NOT NULL) -- Ensure it's a valid datetime
    ,
  ended_at TEXT
    NOT NULL
    CHECK (DATETIME(ended_at) IS NOT NULL) -- Ensure it's a valid datetime
    ,
  kbps_down INTEGER
    NOT NULL
    CHECK(kbps_down > 0)
    ,
  hostname TEXT
    NOT NULL
);
INSERT INTO speedtests_tmp SELECT * FROM speedtests;
DROP TABLE speedtests;
ALTER TABLE speedtests_tmp RENAME TO speedtests;
