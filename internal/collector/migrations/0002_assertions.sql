-- 0002: typed assertion payloads. Exactly one subtype child per assertion is a
-- deferred validation enforced at the publication gate, not a row CHECK.

CREATE TABLE assertions (
    id TEXT PRIMARY KEY,
    observation_id TEXT NOT NULL REFERENCES observations(id) ON DELETE RESTRICT,
    subject_source_id TEXT NOT NULL REFERENCES sources(id) ON DELETE RESTRICT,
    subject_upstream_id TEXT NOT NULL,
    locator TEXT NOT NULL,
    verbatim BYTEA NOT NULL,
    parser_version TEXT NOT NULL,
    evidence_class TEXT NOT NULL CHECK (evidence_class IN
        ('provider_reported','third_party_reported','competitor_reported','host_observed')),
    payload_fingerprint BYTEA NOT NULL CHECK (octet_length(payload_fingerprint) = 32),
    UNIQUE (observation_id, subject_source_id, subject_upstream_id, locator, payload_fingerprint)
);
CREATE INDEX assertions_subject ON assertions (subject_source_id, subject_upstream_id);

CREATE TABLE capability_assertions (
    assertion_id TEXT PRIMARY KEY REFERENCES assertions(id) ON DELETE CASCADE,
    -- capability vocabulary is policy-validated; the spec does not enumerate tokens.
    capability TEXT NOT NULL CHECK (capability <> ''),
    support TEXT NOT NULL CHECK (support IN ('supported','unsupported','unknown')),
    operation TEXT NOT NULL CHECK (operation IN
        ('text_assistance','vision_understanding','image_generation','image_editing')),
    direction TEXT NOT NULL CHECK (direction IN ('input','output','not_applicable')),
    -- Modality assertions require an input/output direction; tool use may use
    -- not_applicable (image-capabilities.md).
    CHECK (capability = 'tool_use' OR direction IN ('input','output'))
);
CREATE INDEX capability_projection ON capability_assertions (capability, support);
CREATE INDEX capability_admission ON capability_assertions (operation, direction, capability, support);

CREATE TABLE image_format_assertions (
    assertion_id TEXT PRIMARY KEY REFERENCES assertions(id) ON DELETE CASCADE,
    operation TEXT NOT NULL CHECK (operation IN
        ('text_assistance','vision_understanding','image_generation','image_editing')),
    role TEXT NOT NULL CHECK (role IN
        ('request','input_image','output_image','reference','mask')),
    -- format is an open vocabulary by design: advertised format names follow
    -- upstream providers (for example webp, png, jpeg) and are validated at
    -- ingestion against the source-advertised set, not a closed enum here.
    format TEXT NOT NULL CHECK (format <> ''),
    CHECK (operation <> 'text_assistance')
);
CREATE INDEX image_format_membership ON image_format_assertions (operation, role, format);

CREATE TABLE image_semantic_assertions (
    assertion_id TEXT PRIMARY KEY REFERENCES assertions(id) ON DELETE CASCADE,
    operation TEXT NOT NULL CHECK (operation = 'image_editing'),
    role TEXT NOT NULL CHECK (role = 'mask'),
    semantic TEXT NOT NULL CHECK (semantic IN
        ('alpha_zero_edits','alpha_one_edits','white_edits','black_edits'))
);
CREATE INDEX image_semantic_lookup ON image_semantic_assertions (operation, role, semantic);

CREATE TABLE image_requirement_assertions (
    assertion_id TEXT PRIMARY KEY REFERENCES assertions(id) ON DELETE CASCADE,
    operation TEXT NOT NULL CHECK (operation IN
        ('text_assistance','vision_understanding','image_generation','image_editing')),
    role TEXT NOT NULL CHECK (role IN ('reference','mask')),
    requirement TEXT NOT NULL CHECK (requirement IN ('required','optional','forbidden','unknown')),
    CHECK (role <> 'mask' OR operation = 'image_editing')
);
CREATE INDEX image_requirement_admission ON image_requirement_assertions (operation, role, requirement);

CREATE TABLE limit_assertions (
    assertion_id TEXT PRIMARY KEY REFERENCES assertions(id) ON DELETE CASCADE,
    limit_name TEXT NOT NULL CHECK (limit_name IN
        ('context_tokens','input_tokens','output_tokens','image_count','reference_image_count',
         'mask_count','image_width','image_height','image_pixels','image_bytes')),
    value BIGINT NOT NULL CHECK (value >= 0),
    unit TEXT NOT NULL CHECK (unit IN ('token','image','pixel','pixel_area','byte')),
    bound TEXT NOT NULL CHECK (bound IN ('minimum','maximum','exact')),
    scope TEXT NOT NULL CHECK (scope IN ('input','output','combined')),
    operation TEXT NOT NULL CHECK (operation IN
        ('text_assistance','vision_understanding','image_generation','image_editing')),
    role TEXT NOT NULL CHECK (role IN
        ('request','input_image','output_image','reference','mask')),
    -- Legal name/unit/scope/operation/role combinations per
    -- image-capabilities.md: generation/editing only for output images,
    -- mask only for editing, input roles unrestricted by operation.
    CHECK (
        (limit_name = 'context_tokens' AND unit = 'token' AND role = 'request')
        OR (limit_name = 'input_tokens' AND unit = 'token' AND scope = 'input' AND role = 'request')
        OR (limit_name = 'output_tokens' AND unit = 'token' AND scope = 'output' AND role = 'request')
        OR (limit_name = 'image_count' AND unit = 'image' AND scope IN ('input','output')
            AND ((scope = 'input' AND role = 'input_image')
                 OR (scope = 'output' AND role = 'output_image'
                     AND operation IN ('image_generation','image_editing'))))
        OR (limit_name = 'reference_image_count' AND unit = 'image' AND scope = 'input' AND role = 'reference')
        OR (limit_name = 'mask_count' AND unit = 'image' AND scope = 'input' AND role = 'mask'
            AND operation = 'image_editing')
        OR (limit_name IN ('image_width','image_height') AND unit = 'pixel' AND
            ((scope = 'input' AND role IN ('input_image','reference','mask')
              AND (role <> 'mask' OR operation = 'image_editing'))
             OR (scope = 'output' AND role = 'output_image'
                 AND operation IN ('image_generation','image_editing'))))
        OR (limit_name = 'image_pixels' AND unit = 'pixel_area' AND
            ((scope = 'input' AND role IN ('input_image','reference','mask')
              AND (role <> 'mask' OR operation = 'image_editing'))
             OR (scope = 'output' AND role = 'output_image'
                 AND operation IN ('image_generation','image_editing'))))
        OR (limit_name = 'image_bytes' AND unit = 'byte' AND
            ((scope = 'input' AND role IN ('input_image','reference','mask')
              AND (role <> 'mask' OR operation = 'image_editing'))
             OR (scope = 'output' AND role = 'output_image'
                 AND operation IN ('image_generation','image_editing'))))
    )
);
CREATE INDEX limit_intersection ON limit_assertions (operation, scope, role, limit_name);

CREATE TABLE claim_assertions (
    assertion_id TEXT PRIMARY KEY REFERENCES assertions(id) ON DELETE CASCADE,
    dimension TEXT NOT NULL CHECK (dimension IN
        ('general','code','review','reasoning','extraction','writing','translation','vision',
         'image_generation','image_editing')),
    polarity TEXT NOT NULL CHECK (polarity IN ('strength','weakness','operational_limit')),
    claim_text TEXT NOT NULL
);

CREATE TABLE assertion_retractions (
    assertion_id TEXT NOT NULL REFERENCES assertions(id) ON DELETE RESTRICT,
    observation_id TEXT NOT NULL REFERENCES observations(id) ON DELETE RESTRICT,
    reason TEXT NOT NULL,
    recorded_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (assertion_id, observation_id)
);
