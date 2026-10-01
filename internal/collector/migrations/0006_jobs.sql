-- 0006: collection jobs, fenced leases, attempts, typed stage outputs and
-- completions. Authority: delivery-storage.md#authority-tables.
-- job_completions.job_id references stage_outputs (one output per job) in
-- addition to the lease tuple, per delivery-storage.

CREATE TABLE jobs (
    id TEXT PRIMARY KEY,
    source_id TEXT NOT NULL REFERENCES sources(id) ON DELETE RESTRICT,
    resource_url TEXT NOT NULL,
    stage TEXT NOT NULL CHECK (stage IN ('fetch','parse','resolve','compile')),
    input_hash BYTEA NOT NULL CHECK (octet_length(input_hash) = 32),
    parser_version TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    attempts BIGINT NOT NULL CHECK (attempts >= 0),
    fencing_token BIGINT NOT NULL CHECK (fencing_token >= 0),
    UNIQUE (source_id, resource_url, stage, input_hash, parser_version)
);
CREATE INDEX jobs_work_selection ON jobs (source_id, stage, created_at);

CREATE TABLE job_leases (
    job_id TEXT NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
    token BIGINT NOT NULL CHECK (token > 0),
    worker_id TEXT NOT NULL,
    leased_at TIMESTAMPTZ NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (job_id, token),
    CHECK (expires_at > leased_at)
);
CREATE INDEX job_lease_recovery ON job_leases (expires_at);

CREATE TABLE job_attempts (
    job_id TEXT NOT NULL,
    token BIGINT NOT NULL,
    observed_at TIMESTAMPTZ NOT NULL,
    outcome TEXT NOT NULL CHECK (outcome IN
        ('observed','deferred','quarantined','failed_terminal')),
    detail TEXT NOT NULL,
    retry_not_before TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (job_id, token, observed_at),
    FOREIGN KEY (job_id, token) REFERENCES job_leases(job_id, token) ON DELETE CASCADE,
    CHECK (outcome <> 'deferred' OR retry_not_before > observed_at),
    CHECK (outcome = 'deferred' OR retry_not_before = observed_at)
);
CREATE INDEX job_retry_sweep ON job_attempts (retry_not_before, outcome);

CREATE TABLE stage_outputs (
    job_id TEXT PRIMARY KEY REFERENCES jobs(id) ON DELETE CASCADE,
    output_id TEXT NOT NULL,
    output_hash BYTEA NOT NULL CHECK (octet_length(output_hash) = 32),
    output_bytes BYTEA NOT NULL,
    recorded_at TIMESTAMPTZ NOT NULL
);

-- observation_id nullability is the one justified nullable column: a compile
-- output has no observation fact.
CREATE TABLE stage_output_links (
    observation_id TEXT REFERENCES observations(id) ON DELETE RESTRICT,
    job_id TEXT PRIMARY KEY REFERENCES stage_outputs(job_id) ON DELETE CASCADE,
    source_id TEXT NOT NULL REFERENCES sources(id) ON DELETE RESTRICT,
    resource_url TEXT NOT NULL,
    stage TEXT NOT NULL CHECK (stage IN ('fetch','parse','resolve','compile')),
    input_hash BYTEA NOT NULL CHECK (octet_length(input_hash) = 32),
    parser_version TEXT NOT NULL,
    output_id TEXT NOT NULL,
    output_hash BYTEA NOT NULL CHECK (octet_length(output_hash) = 32),
    recorded_at TIMESTAMPTZ NOT NULL,
    UNIQUE (job_id, output_id, output_hash)
);
CREATE INDEX stage_output_by_observation ON stage_output_links (observation_id);

CREATE TABLE job_completions (
    job_id TEXT PRIMARY KEY REFERENCES stage_outputs(job_id) ON DELETE CASCADE,
    token BIGINT NOT NULL,
    output_id TEXT NOT NULL,
    completed_at TIMESTAMPTZ NOT NULL,
    FOREIGN KEY (job_id, token) REFERENCES job_leases(job_id, token) ON DELETE RESTRICT
);
