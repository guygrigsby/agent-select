-- 0010: snapshot freshness confirmations (delivery-storage V8, ar-2pe.2).
-- Append-only infrastructure fact: the successor pipeline checked the channel's
-- current artifact at confirmed_at and the projection content was unchanged.
CREATE TABLE snapshot_freshness_confirmations (
    artifact_hash BYTEA NOT NULL REFERENCES snapshots (artifact_hash),
    channel       TEXT NOT NULL REFERENCES publication_channels (channel),
    confirmed_at  TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (artifact_hash, channel, confirmed_at)
);

CREATE INDEX snapshot_freshness_latest
    ON snapshot_freshness_confirmations (channel, confirmed_at DESC);
