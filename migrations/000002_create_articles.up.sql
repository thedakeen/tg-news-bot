CREATE TABLE IF NOT EXISTS articles(
    id BIGSERIAL PRIMARY KEY,
    source_id INT NOT NULL,
    title VARCHAR(255) NOT NULL,
    link VARCHAR(255) NOT NULL,
    summary TEXT NOT NULL,
    published_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL,
    posted_at TIMESTAMP,
    CONSTRAINT fk_articles_source_id FOREIGN KEY (source_id) REFERENCES sources(id)
);