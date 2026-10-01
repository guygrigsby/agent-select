-- 0004: benchmark runs, measurements and cohort policy. Authority:
-- benchmark-comparability.md#relational-contract.

CREATE TABLE benchmark_runs (
    id TEXT PRIMARY KEY,
    observation_id TEXT NOT NULL REFERENCES observations(id) ON DELETE RESTRICT,
    locator TEXT NOT NULL,
    identity_kind TEXT NOT NULL CHECK (identity_kind IN ('publisher_run','source_record')),
    identity_value TEXT NOT NULL,
    content_hash BYTEA NOT NULL CHECK (octet_length(content_hash) = 32),
    UNIQUE (observation_id, locator, identity_kind, identity_value)
);
CREATE INDEX benchmark_run_resolution ON benchmark_runs (identity_kind, identity_value);

CREATE TABLE benchmark_measurements (
    assertion_id TEXT PRIMARY KEY REFERENCES assertions(id) ON DELETE CASCADE,
    run_id TEXT NOT NULL REFERENCES benchmark_runs(id) ON DELETE RESTRICT,
    benchmark_label TEXT NOT NULL,
    metric_label TEXT NOT NULL,
    raw_value NUMERIC NOT NULL,
    unit TEXT NOT NULL,
    direction TEXT NOT NULL CHECK (direction IN ('higher','lower','unordered'))
);
CREATE INDEX benchmark_raw_grouping
    ON benchmark_measurements (benchmark_label, metric_label, unit, direction);

CREATE TABLE measurement_revision_facts (
    assertion_id TEXT NOT NULL REFERENCES benchmark_measurements(assertion_id) ON DELETE CASCADE,
    revision_id TEXT NOT NULL REFERENCES model_revisions(id) ON DELETE RESTRICT,
    identity_evidence_id TEXT NOT NULL REFERENCES assertions(id) ON DELETE RESTRICT,
    PRIMARY KEY (assertion_id, revision_id, identity_evidence_id)
);
CREATE INDEX measurement_revision_provenance ON measurement_revision_facts (revision_id);

CREATE TABLE benchmark_metadata (
    assertion_id TEXT NOT NULL REFERENCES benchmark_measurements(assertion_id) ON DELETE CASCADE,
    component TEXT NOT NULL,
    knowledge TEXT NOT NULL CHECK (knowledge IN ('known','unknown','not_applicable')),
    PRIMARY KEY (assertion_id, component)
);
CREATE INDEX benchmark_metadata_missing ON benchmark_metadata (component, knowledge);

CREATE TABLE benchmark_known_values (
    assertion_id TEXT NOT NULL,
    component TEXT NOT NULL,
    canonical_value TEXT NOT NULL,
    evidence_locator TEXT NOT NULL,
    PRIMARY KEY (assertion_id, component),
    FOREIGN KEY (assertion_id, component)
        REFERENCES benchmark_metadata(assertion_id, component) ON DELETE CASCADE,
    CHECK (canonical_value <> '')
);

CREATE TABLE benchmark_unknown_reasons (
    assertion_id TEXT NOT NULL,
    component TEXT NOT NULL,
    reason TEXT NOT NULL,
    PRIMARY KEY (assertion_id, component),
    FOREIGN KEY (assertion_id, component)
        REFERENCES benchmark_metadata(assertion_id, component) ON DELETE CASCADE,
    CHECK (reason <> '')
);

CREATE TABLE benchmark_not_applicable_facts (
    assertion_id TEXT NOT NULL,
    component TEXT NOT NULL,
    rationale TEXT NOT NULL,
    evidence_locator TEXT NOT NULL,
    PRIMARY KEY (assertion_id, component),
    FOREIGN KEY (assertion_id, component)
        REFERENCES benchmark_metadata(assertion_id, component) ON DELETE CASCADE
);

CREATE TABLE benchmark_run_periods (
    run_id TEXT PRIMARY KEY REFERENCES benchmark_runs(id) ON DELETE CASCADE,
    started_at TIMESTAMPTZ NOT NULL,
    ended_at TIMESTAMPTZ NOT NULL,
    evidence_locator TEXT NOT NULL,
    CHECK (started_at <= ended_at)
);
CREATE INDEX benchmark_period_window ON benchmark_run_periods (started_at, ended_at);

CREATE TABLE benchmark_cohort_policies (
    version TEXT PRIMARY KEY,
    family TEXT NOT NULL CHECK (family <> ''),
    canonicalization_version TEXT NOT NULL,
    content_hash BYTEA NOT NULL UNIQUE CHECK (octet_length(content_hash) = 32),
    recorded_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE benchmark_cohort_rules (
    policy_version TEXT NOT NULL REFERENCES benchmark_cohort_policies(version) ON DELETE CASCADE,
    component TEXT NOT NULL,
    value_type TEXT NOT NULL CHECK (value_type IN
        ('integer_count','bounded_decimal','artifact_id','closed_enum','instant_pair')),
    canonical_grammar TEXT NOT NULL,
    critical BOOLEAN NOT NULL,
    not_applicable_allowed BOOLEAN NOT NULL,
    rationale TEXT NOT NULL,
    PRIMARY KEY (policy_version, component)
);

CREATE TABLE benchmark_cohort_memberships (
    assertion_id TEXT NOT NULL REFERENCES benchmark_measurements(assertion_id) ON DELETE CASCADE,
    policy_version TEXT NOT NULL REFERENCES benchmark_cohort_policies(version) ON DELETE RESTRICT,
    cohort_hash BYTEA NOT NULL CHECK (octet_length(cohort_hash) = 32),
    PRIMARY KEY (assertion_id, policy_version)
);
CREATE INDEX cohort_candidate_lookup
    ON benchmark_cohort_memberships (policy_version, cohort_hash);

CREATE TABLE measurement_sample_counts (
    assertion_id TEXT PRIMARY KEY REFERENCES benchmark_measurements(assertion_id) ON DELETE CASCADE,
    count BIGINT NOT NULL CHECK (count > 0),
    evidence_locator TEXT NOT NULL
);

CREATE TABLE measurement_intervals (
    assertion_id TEXT NOT NULL REFERENCES benchmark_measurements(assertion_id) ON DELETE CASCADE,
    method TEXT NOT NULL,
    confidence NUMERIC NOT NULL CHECK (confidence > 0 AND confidence < 1),
    lower_bound NUMERIC NOT NULL,
    upper_bound NUMERIC NOT NULL,
    evidence_locator TEXT NOT NULL,
    PRIMARY KEY (assertion_id, method, confidence),
    CHECK (lower_bound <= upper_bound)
);

CREATE TABLE measurement_standard_errors (
    assertion_id TEXT NOT NULL REFERENCES benchmark_measurements(assertion_id) ON DELETE CASCADE,
    method TEXT NOT NULL,
    standard_error NUMERIC NOT NULL CHECK (standard_error >= 0),
    evidence_locator TEXT NOT NULL,
    PRIMARY KEY (assertion_id, method)
);
