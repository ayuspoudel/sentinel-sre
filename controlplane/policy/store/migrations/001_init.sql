CREATE TABLE IF NOT EXISTS policies (
    name TEXT PRIMARY KEY,
    spec JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS policy_status (
    policy_name TEXT PRIMARY KEY,
    status JSONB NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT fk_policy
        FOREIGN KEY(policy_name)
        REFERENCES policies(name)
        ON DELETE CASCADE
);
