CREATE TABLE IF NOT EXISTS courses (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    description TEXT NOT NULL,
    price       TEXT NOT NULL,
    user_id     TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_id ON courses(id);
