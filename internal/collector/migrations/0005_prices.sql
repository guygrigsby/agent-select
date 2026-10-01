-- 0005: price assertions and scope/proof children. Authority:
-- contracts.md#data-layer, domain-model.md PriceComponent.
-- price_assertions.category and price_bound_proofs.component vocabularies are
-- validated at ingestion; the specs do not enumerate their tokens.

CREATE TABLE price_assertions (
    assertion_id TEXT PRIMARY KEY REFERENCES assertions(id) ON DELETE CASCADE,
    offer_id TEXT NOT NULL REFERENCES offers(id) ON DELETE RESTRICT,
    currency TEXT NOT NULL CHECK (char_length(currency) = 3),
    amount NUMERIC NOT NULL CHECK (amount >= 0),
    unit TEXT NOT NULL CHECK (unit IN
        ('token','million_tokens','image','megapixel','second','character','request')),
    category TEXT NOT NULL CHECK (category <> ''),
    quote_knowledge TEXT NOT NULL CHECK (quote_knowledge IN
        ('exact','proven_upper_bound','incomplete')),
    scope_hash BYTEA NOT NULL CHECK (octet_length(scope_hash) = 32)
);
CREATE INDEX price_quote_reconciliation ON price_assertions (currency, category, scope_hash);

CREATE TABLE price_scope_items (
    assertion_id TEXT NOT NULL REFERENCES price_assertions(assertion_id) ON DELETE CASCADE,
    dimension TEXT NOT NULL CHECK (dimension IN
        ('direction','provider','region','service_tier','context_band','cache_retention',
         'operation','width','height','image_count','quality','style','steps')),
    knowledge TEXT NOT NULL CHECK (knowledge IN ('known','unknown','not_applicable')),
    value_bytes BYTEA NOT NULL,
    PRIMARY KEY (assertion_id, dimension),
    CHECK (
        (knowledge = 'known' AND octet_length(value_bytes) > 0)
        OR (knowledge <> 'known' AND octet_length(value_bytes) = 0)
    )
);

CREATE TABLE price_bound_proofs (
    assertion_id TEXT NOT NULL REFERENCES price_assertions(assertion_id) ON DELETE CASCADE,
    component TEXT NOT NULL CHECK (component <> ''),
    covered_range_bytes BYTEA NOT NULL,
    evidence_assertion_id TEXT NOT NULL REFERENCES assertions(id) ON DELETE RESTRICT,
    PRIMARY KEY (assertion_id, component)
);
