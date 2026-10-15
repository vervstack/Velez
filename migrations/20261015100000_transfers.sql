-- +goose Up
-- +goose StatementBegin

CREATE TYPE velez.transfer_status AS ENUM ('PENDING', 'RUNNING', 'FAILED', 'DONE', 'CANCELLED');

CREATE TABLE velez.transfer_tokens
(
    id         UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    token_hash TEXT        NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at    TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON COLUMN velez.transfer_tokens.used_at IS 'NULL until consumed by the Transfer stream Begin frame';

CREATE TABLE velez.transfers
(
    id             UUID PRIMARY KEY               DEFAULT gen_random_uuid(),
    is_source      BOOLEAN               NOT NULL,
    container_name TEXT                  NOT NULL,
    peer_address   TEXT                  NOT NULL DEFAULT '',
    status         velez.transfer_status NOT NULL DEFAULT 'PENDING',
    error          TEXT                  NOT NULL DEFAULT '',
    total_bytes    BIGINT                NOT NULL DEFAULT 0,
    done_bytes     BIGINT                NOT NULL DEFAULT 0,
    created_at     TIMESTAMPTZ           NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ           NOT NULL DEFAULT NOW()
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE velez.transfers;

DROP TABLE velez.transfer_tokens;

DROP TYPE velez.transfer_status;

-- +goose StatementEnd
