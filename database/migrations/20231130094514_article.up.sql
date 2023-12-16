CREATE TABLE article(
    id uuid PRIMARY KEY,
    slug text NOT NULL UNIQUE,
    title text NOT NULL,
    description text NOT NULL,
    body text NOT NULL,
    author_id uuid NOT NULL,
    created_at timestamptz NOT NULL DEFAULT (now()),
    updated_at timestamptz NOT NULL DEFAULT (now()),
    FOREIGN KEY (author_id) REFERENCES appuser(id) ON DELETE CASCADE ON UPDATE CASCADE
);

-- create index for author_id
CREATE INDEX article_author_id_idx ON article(author_id);

CREATE TABLE comment(
    id int PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    body text NOT NULL,
    article_id uuid NOT NULL,
    author_id uuid NOT NULL,
    created_at timestamptz NOT NULL DEFAULT (now()),
    updated_at timestamptz NOT NULL DEFAULT (now()),
    FOREIGN KEY (article_id) REFERENCES article(id) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (author_id) REFERENCES appuser(id) ON DELETE CASCADE ON UPDATE CASCADE
);

-- create index for article_id
CREATE INDEX comment_article_id_idx ON comment(article_id);

-- create index for author_id
CREATE INDEX comment_author_id_idx ON comment(author_id);

CREATE TABLE tag(
    id uuid PRIMARY KEY,
    name varchar NOT NULL UNIQUE,
    created_at timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE article_tag(
    article_id uuid NOT NULL,
    tag_id uuid NOT NULL,
    created_at timestamptz NOT NULL DEFAULT (now()),
    PRIMARY KEY (article_id, tag_id),
    FOREIGN KEY (article_id) REFERENCES article(id) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (tag_id) REFERENCES tag(id) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE article_favorite(
    appuser_id uuid NOT NULL,
    article_id uuid NOT NULL,
    created_at timestamptz NOT NULL DEFAULT (now()),
    PRIMARY KEY (appuser_id, article_id),
    FOREIGN KEY (appuser_id) REFERENCES appuser(id) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (article_id) REFERENCES article(id) ON DELETE CASCADE ON UPDATE CASCADE
);
