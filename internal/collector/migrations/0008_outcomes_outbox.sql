-- 0008: host outcome facts and the transactional outbox. Authority:
-- delivery-storage.md#authority-tables.

CREATE TABLE outcome_facts (
    receipt_id TEXT PRIMARY KEY,
    observation_id TEXT NOT NULL REFERENCES observations(id) ON DELETE RESTRICT,
    request_id TEXT NOT NULL CHECK (request_id <> ''),
    offer_id TEXT NOT NULL,
    task TEXT NOT NULL CHECK (task IN
        ('general','code','review','reasoning','extraction','writing','translation','vision',
         'image_generation','image_editing')),
    outcome TEXT NOT NULL CHECK (outcome IN
        ('validated_success','validated_failure','unvalidated_completion','indeterminate',
         'cancelled')),
    recorded_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX outcome_projection ON outcome_facts (offer_id, recorded_at DESC);
CREATE INDEX outcome_signal_derivation ON outcome_facts (task, outcome);

-- Validation is a conditionally-present append-only child fact. The trigger
-- rejects a child row whose parent outcome is not validated; the record_receipt
-- port writes both rows in one transaction and the validation sweep re-verifies
-- that no validated parent lacks a child row.
CREATE TABLE validated_outcomes (
    receipt_id TEXT PRIMARY KEY REFERENCES outcome_facts(receipt_id) ON DELETE CASCADE,
    validator_version TEXT NOT NULL CHECK (validator_version <> ''),
    recorded_at TIMESTAMPTZ NOT NULL
);

CREATE FUNCTION validated_outcome_guard() RETURNS trigger AS $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM outcome_facts f
        WHERE f.receipt_id = NEW.receipt_id
          AND f.outcome IN ('validated_success','validated_failure')
    ) THEN
        RAISE EXCEPTION 'validated_outcomes requires a validated parent outcome for %',
            NEW.receipt_id;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER validated_outcome_guard
    BEFORE INSERT OR UPDATE ON validated_outcomes
    FOR EACH ROW EXECUTE FUNCTION validated_outcome_guard();

-- The seven usage components per receipt split by knowledge; the disjoint and
-- exhaustive union is enforced by record_receipt in-transaction and re-verified
-- by the validation sweep.
CREATE TABLE outcome_usage_known (
    receipt_id TEXT NOT NULL REFERENCES outcome_facts(receipt_id) ON DELETE CASCADE,
    component TEXT NOT NULL CHECK (component IN
        ('input_tokens','output_tokens','input_images','output_images','references','searches',
         'elapsed_millis')),
    value BIGINT NOT NULL CHECK (value >= 0),
    PRIMARY KEY (receipt_id, component)
);

CREATE TABLE outcome_usage_unknown (
    receipt_id TEXT NOT NULL REFERENCES outcome_facts(receipt_id) ON DELETE CASCADE,
    component TEXT NOT NULL CHECK (component IN
        ('input_tokens','output_tokens','input_images','output_images','references','searches',
         'elapsed_millis')),
    PRIMARY KEY (receipt_id, component)
);

CREATE TABLE outbox_events (
    event_id TEXT PRIMARY KEY,
    aggregate_id TEXT NOT NULL,
    aggregate_sequence BIGINT NOT NULL CHECK (aggregate_sequence > 0),
    event_name TEXT NOT NULL CHECK (event_name IN
        ('SourcePolicyChanged','ObservationRecorded','StageOutputRecorded','JobDeferred',
         'JobCompleted','RevisionEstablished','OfferListingChanged','SnapshotCompiled',
         'SnapshotValidated','SnapshotPublished','SnapshotRetired','ReceiptRecorded',
         'IdentityBindingEstablished','IdentityBindingConfirmed','IdentityBindingEnded',
         'IdentityBindingSuperseded','IdentityPolicyAvailable')),
    schema_version INTEGER NOT NULL CHECK (schema_version > 0),
    occurred_at TIMESTAMPTZ NOT NULL,
    payload_hash BYTEA NOT NULL CHECK (octet_length(payload_hash) = 32),
    payload BYTEA NOT NULL,
    UNIQUE (aggregate_id, aggregate_sequence)
);
CREATE INDEX outbox_delivery_scan ON outbox_events (occurred_at, event_id);

CREATE TABLE consumer_receipts (
    consumer_id TEXT NOT NULL,
    event_id TEXT NOT NULL REFERENCES outbox_events(event_id) ON DELETE RESTRICT,
    received_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (consumer_id, event_id)
);
