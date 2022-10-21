#!/bin/env bash

# Assuming you just did `cd .. && make dev/docker-db`
echo "====> CREATING todos TABLE"
psql "postgres://postgres:postgres@127.0.0.1:5432/todos?sslmode=disable" <<EOF
CREATE TABLE IF NOT EXISTS todos (
  id SERIAL PRIMARY KEY,
  title TEXT,
  done BOOL
);
EOF

# Assuming you just did: `vault server -dev`
export VAULT_ADDR='http://127.0.0.1:8200'

echo "====> ENABLING APP ROLE"
vault auth enable approle
echo 'path "secret/*" { capabilities = ["read"] }
      path "database/creds/todos-role" { capabilities = ["read"] }
' | vault policy write get-da-secrets -

vault write auth/approle/role/get-da-secrets policies="get-da-secrets"

echo "====> ENABLING POSTGRES ENGINE"
vault secrets enable database
vault write database/config/todos-database \
      plugin_name="postgresql-database-plugin" \
      allowed_roles="todos-role" \
      connection_url="postgresql://{{username}}:{{password}}@localhost:5432/todos?sslmode=disable" \
      username="postgres" \
      password="postgres"

vault write database/roles/todos-role \
      db_name="todos-database" \
      creation_statements="CREATE ROLE \"{{name}}\" WITH LOGIN PASSWORD '{{password}}' VALID UNTIL '{{expiration}}'; \
        GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO \"{{name}}\"; \
        GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO \"{{name}}\";" \
      default_ttl="24h" \
      max_ttl="48h"

echo "====> ENVIRONMENT TO USE"

echo export VAULT_ADDR=${VAULT_ADDR}
echo export APPROLE_ROLE_ID="$(vault read -field=role_id auth/approle/role/get-da-secrets/role-id)"
echo export APPROLE_SECRET_ID="$(vault write -f -field=secret_id auth/approle/role/get-da-secrets/secret-id)"
echo export APPROLE_PATH="approle"
echo
