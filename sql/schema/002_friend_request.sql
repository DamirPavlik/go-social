-- +goose Up
CREATE TABLE friend_request (
    id SERIAL PRIMARY KEY,
    sender_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reciever_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status TEXT CHECK (status IN ('pending', 'accepted', 'declined')) DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT now(),
    UNIQUE(sender_id, reciever_id)
);

-- +goose Down
DROP TABLE friend_request;