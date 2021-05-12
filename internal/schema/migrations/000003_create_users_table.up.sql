CREATE TABLE users
(
    user_id       UUID,
    name          TEXT,
    email         TEXT UNIQUE,
    active		  BOOLEAN DEFAULT TRUE,
    roles         TEXT[],
    password_hash TEXT,
    date_created  TIMESTAMP WITH TIME ZONE,
    date_updated  TIMESTAMP WITH TIME ZONE,
    PRIMARY KEY (user_id)
);
