-- 0001: sources, policies, content store and observations.
-- Authority: docs/specs/contracts.md#data-layer, docs/specs/delivery-storage.md#authority-tables.
-- All hashes are 32-byte SHA-256 values. Evidence rows are append-only.

CREATE TABLE sources (
    id TEXT PRIMARY KEY,
    url TEXT NOT NULL UNIQUE,
    source_kind TEXT NOT NULL CHECK (source_kind IN
        ('vercel_catalog','modelsdev','provider_api','provider_site','eval_registry','host_local')),
    parser_version TEXT NOT NULL,
    freshness_ceiling_seconds BIGINT NOT NULL CHECK (freshness_ceiling_seconds > 0),
    enabled BOOLEAN NOT NULL
);
CREATE INDEX sources_active ON sources (enabled, source_kind);

CREATE TABLE source_policies (
    source_id TEXT NOT NULL REFERENCES sources(id) ON DELETE RESTRICT,
    config_hash BYTEA NOT NULL CHECK (octet_length(config_hash) = 32),
    research TEXT NOT NULL CHECK (research IN ('allowed','denied','unknown')),
    extraction TEXT NOT NULL CHECK (extraction IN ('allowed','denied','unknown')),
    redistribution TEXT NOT NULL CHECK (redistribution IN ('allowed','denied','unknown')),
    terms_url TEXT NOT NULL,
    license_text TEXT NOT NULL,
    checked_at TIMESTAMPTZ NOT NULL,
    effective_at TIMESTAMPTZ NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (source_id, config_hash),
    CHECK (effective_at < expires_at)
);
CREATE INDEX source_policies_current ON source_policies (source_id, effective_at DESC);

CREATE TABLE content_blobs (
    hash BYTEA PRIMARY KEY CHECK (octet_length(hash) = 32),
    bytes BYTEA NOT NULL
);

CREATE TABLE observations (
    id TEXT PRIMARY KEY,
    source_id TEXT NOT NULL REFERENCES sources(id) ON DELETE RESTRICT,
    resource_url TEXT NOT NULL,
    retrieved_at TIMESTAMPTZ NOT NULL,
    http_status INTEGER NOT NULL CHECK (http_status BETWEEN 0 AND 599),
    outcome TEXT NOT NULL CHECK (outcome IN
        ('success','unchanged','http_error','source_blocked','invalid_shape','quarantine')),
    body_hash BYTEA NOT NULL REFERENCES content_blobs(hash) ON DELETE RESTRICT
        CHECK (octet_length(body_hash) = 32),
    headers_bytes BYTEA NOT NULL,
    collector_version TEXT NOT NULL,
    -- Delivery-storage outcome/status pairing, column-only and subquery-free.
    -- Status 0 with success is legitimate only for host_local receipt observations
    -- (no HTTP exchange); fetch ports never write it and the validation sweep
    -- re-verifies the origin constraint.
    CHECK (
        (outcome = 'success' AND http_status IN (0, 200)) OR
        (outcome = 'unchanged' AND http_status IN (200, 304)) OR
        (outcome = 'http_error' AND (http_status = 0 OR http_status >= 400)) OR
        (outcome = 'source_blocked' AND http_status IN (401, 403, 429, 451)) OR
        (outcome = 'invalid_shape' AND http_status BETWEEN 200 AND 299) OR
        (outcome = 'quarantine')
    )
);
CREATE INDEX observations_latest ON observations (source_id, resource_url, retrieved_at DESC);
CREATE INDEX observations_body ON observations (body_hash);

CREATE TABLE unchanged_observations (
    observation_id TEXT PRIMARY KEY REFERENCES observations(id) ON DELETE CASCADE,
    prior_observation_id TEXT NOT NULL REFERENCES observations(id) ON DELETE RESTRICT,
    CHECK (observation_id <> prior_observation_id)
);
