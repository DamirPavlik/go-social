-- +goose Up
ALTER TABLE users ADD COLUMN email TEXT UNIQUE;

-- +goose Down 
ALTER TABLE users DROP COLUMN email TEXT UNIQUE;