-- +goose Up
CREATE TABLE avatars (
                         id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                         user_id BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
                         file_name VARCHAR(255) NOT NULL,
                         mime_type VARCHAR(100) NOT NULL,
                         size_bytes BIGINT NOT NULL,
                         s3_key VARCHAR(500) NOT NULL,
                         thumbnail_s3_keys JSONB,
                         upload_status VARCHAR(50) DEFAULT 'uploading',
                         processing_status VARCHAR(50) DEFAULT 'pending',
                         created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
                         updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
                         deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_avatars_user_id
    ON avatars(user_id)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_avatars_status
    ON avatars(upload_status, processing_status);

-- +goose Down
DROP TABLE avatars;