CREATE TABLE recipes (
    id              SERIAL PRIMARY KEY,
    name            TEXT,
    author          TEXT,
    budget          TEXT,
    difficulty      TEXT,
    setup_time      TEXT,
    cook_time       TEXT,
    total_time      TEXT,
    people_quantity TEXT
);

CREATE TABLE ingredients (
    id              SERIAL PRIMARY KEY,
    recipe_id       INT REFERENCES recipes(id),
    index           INT,
    ingredient      TEXT
);

CREATE TABLE tags (
    id              SERIAL PRIMARY KEY,
    recipe_id       INT REFERENCES recipes(id),
    index           INT,
    tag             TEXT
);
