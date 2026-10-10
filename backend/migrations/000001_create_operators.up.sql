-- operators: the human accounts that can sign in. Provision the first one with
-- cmd/provision-operator.
CREATE TABLE operators (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    email text NOT NULL UNIQUE,
    password_hash text NOT NULL
);
