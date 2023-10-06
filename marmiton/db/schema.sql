CREATE TABLE recipes (
    id              SERIAL  PRIMARY KEY,
    name            TEXT,
    author          TEXT,
    tags            TEXT [],
    budget          TEXT,
    difficulty      TEXT,
    setup_time      TEXT,
    cook_time       TEXT,
    total_time      TEXT,
    ingredients     TEXT [],
    people_quantity TEXT
);
