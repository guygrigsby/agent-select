-- 0007: projection policies, snapshot family, publication channels and rights
-- audits. Authority: delivery-storage.md#authority-tables.
-- snapshot_family columns use artifact_hash (delivery-storage supersedes the
-- earlier snapshot_hash spelling in contracts.md).

CREATE TABLE projection_policies (
    version TEXT PRIMARY KEY,
    content_hash BYTEA NOT NULL UNIQUE CHECK (octet_length(content_hash) = 32),
    content_bytes BYTEA NOT NULL,
    recorded_at TIMESTAMPTZ NOT NULL,
    CHECK (version <> '')
);

CREATE TABLE snapshots (
    artifact_hash BYTEA PRIMARY KEY CHECK (octet_length(artifact_hash) = 32),
    schema_version INTEGER NOT NULL CHECK (schema_version = 1),
    policy TEXT NOT NULL CHECK (policy = 'educated-v1'),
    distribution TEXT NOT NULL CHECK (distribution IN ('local','distributable')),
    projection_policy_version TEXT NOT NULL REFERENCES projection_policies(version) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL,
    knowledge_cutoff TIMESTAMPTZ NOT NULL,
    artifact_bytes BYTEA NOT NULL,
    input_manifest_hash BYTEA NOT NULL CHECK (octet_length(input_manifest_hash) = 32),
    coverage_digest BYTEA NOT NULL CHECK (octet_length(coverage_digest) = 32)
);
CREATE INDEX snapshots_retention ON snapshots (created_at DESC);
CREATE INDEX snapshots_compile_replay
    ON snapshots (input_manifest_hash, distribution, projection_policy_version,
                  created_at, knowledge_cutoff);

CREATE TABLE snapshot_inputs (
    artifact_hash BYTEA NOT NULL REFERENCES snapshots(artifact_hash) ON DELETE CASCADE,
    ordinal INTEGER NOT NULL CHECK (ordinal >= 0),
    observation_id TEXT NOT NULL REFERENCES observations(id) ON DELETE RESTRICT,
    PRIMARY KEY (artifact_hash, ordinal),
    UNIQUE (artifact_hash, observation_id)
);

CREATE TABLE snapshot_rights_audits (
    artifact_hash BYTEA NOT NULL REFERENCES snapshots(artifact_hash) ON DELETE CASCADE,
    source_id TEXT NOT NULL REFERENCES sources(id) ON DELETE RESTRICT,
    config_hash BYTEA NOT NULL,
    research TEXT NOT NULL CHECK (research IN ('allowed','denied','unknown')),
    extraction TEXT NOT NULL CHECK (extraction IN ('allowed','denied','unknown')),
    redistribution TEXT NOT NULL CHECK (redistribution IN ('allowed','denied','unknown')),
    decision TEXT NOT NULL CHECK (decision IN ('allowed','denied')),
    decided_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (artifact_hash, source_id),
    FOREIGN KEY (source_id, config_hash)
        REFERENCES source_policies(source_id, config_hash) ON DELETE RESTRICT
);

CREATE TABLE snapshot_validations (
    artifact_hash BYTEA NOT NULL REFERENCES snapshots(artifact_hash) ON DELETE CASCADE,
    validator_version TEXT NOT NULL CHECK (validator_version <> ''),
    outcome TEXT NOT NULL CHECK (outcome IN ('accepted','rejected')),
    findings_bytes BYTEA NOT NULL,
    validated_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (artifact_hash, validator_version)
);

CREATE TABLE publication_channels (
    channel TEXT PRIMARY KEY,
    channel_class TEXT NOT NULL CHECK (channel_class IN ('local','distributable')),
    created_at TIMESTAMPTZ NOT NULL
);

-- Retirement is an append-only lifecycle child row, not a nullable column.
CREATE TABLE publication_channel_retirements (
    channel TEXT PRIMARY KEY REFERENCES publication_channels(channel) ON DELETE CASCADE,
    retired_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE snapshot_publications (
    id TEXT PRIMARY KEY,
    artifact_hash BYTEA NOT NULL REFERENCES snapshots(artifact_hash) ON DELETE RESTRICT,
    channel TEXT NOT NULL REFERENCES publication_channels(channel) ON DELETE RESTRICT,
    channel_class TEXT NOT NULL CHECK (channel_class IN ('local','distributable')),
    rights_decision TEXT NOT NULL CHECK (rights_decision IN ('allowed','denied')),
    artifact_locator TEXT NOT NULL,
    published_at TIMESTAMPTZ NOT NULL,
    UNIQUE (artifact_hash, channel)
);
CREATE INDEX snapshot_active_publication
    ON snapshot_publications (channel, published_at DESC);

-- One BEFORE INSERT trigger rejects both a class mismatch and a retired
-- channel (delivery-storage).
CREATE FUNCTION snapshot_publication_guard() RETURNS trigger AS $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM publication_channels c
        WHERE c.channel = NEW.channel AND c.channel_class <> NEW.channel_class
    ) THEN
        RAISE EXCEPTION 'publication channel class mismatch for %', NEW.channel;
    END IF;
    IF EXISTS (
        SELECT 1 FROM publication_channel_retirements r WHERE r.channel = NEW.channel
    ) THEN
        RAISE EXCEPTION 'publication channel % is retired', NEW.channel;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER snapshot_publication_guard
    BEFORE INSERT ON snapshot_publications
    FOR EACH ROW EXECUTE FUNCTION snapshot_publication_guard();

CREATE TABLE snapshot_retirements (
    artifact_hash BYTEA NOT NULL REFERENCES snapshots(artifact_hash) ON DELETE RESTRICT,
    successor_hash BYTEA NOT NULL REFERENCES snapshots(artifact_hash) ON DELETE RESTRICT,
    publication_id TEXT NOT NULL REFERENCES snapshot_publications(id) ON DELETE RESTRICT,
    retired_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (artifact_hash, publication_id),
    CHECK (artifact_hash <> successor_hash)
);
