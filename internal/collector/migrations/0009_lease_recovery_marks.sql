-- 0009: durable recovery marks on leases. recover_job is idempotent by
-- expired token (contracts.md): the first recovery marks the lease row;
-- replays derive state without mutation, so repeated reconciliation sweeps
-- cannot inflate jobs.attempts. ExpiredLeases excludes marked leases
-- (the job_lease_recovery index on expires_at already serves that scan).
ALTER TABLE job_leases ADD COLUMN recovered_at TIMESTAMPTZ;
