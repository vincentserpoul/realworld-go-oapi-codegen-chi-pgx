CREATE TABLE appuser(
    id uuid PRIMARY KEY,
    email varchar NOT NULL UNIQUE,
    username varchar NOT NULL UNIQUE,
    pwd varchar NOT NULL,
    img text NOT NULL DEFAULT 'https://avatars.githubusercontent.com/u/32737308?v=4',
    bio text NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT (now()),
    updated_at timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE appuser_follows(
    follower_id uuid NOT NULL,
    followee_id uuid NOT NULL,
    created_at timestamptz NOT NULL DEFAULT (now()),
    PRIMARY KEY (follower_id, followee_id),
    FOREIGN KEY (follower_id) REFERENCES appuser(id) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (followee_id) REFERENCES appuser(id) ON DELETE CASCADE ON UPDATE CASCADE
);
