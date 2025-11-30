-- +goose Up
CREATE INDEX idx_comments_post_id ON comments(post_id);
CREATE INDEX idx_comments_parent_id ON comments(parent_id);
CREATE INDEX idx_comments_created_at ON comments(created_at);
CREATE INDEX idx_posts_created_at ON posts(created_at);

-- +goose Down
DROP INDEX idx_comments_post_id;
DROP INDEX idx_comments_parent_id;
DROP INDEX idx_comments_created_at;
DROP INDEX idx_posts_created_at;