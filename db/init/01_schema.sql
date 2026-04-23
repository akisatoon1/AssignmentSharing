CREATE TABLE message (
    id SERIAL PRIMARY KEY,
    content TEXT NOT NULL
);

INSERT INTO message (content) VALUES ('Hello from PostgreSQL!');
