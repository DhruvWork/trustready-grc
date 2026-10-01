#!/bin/bash
# Copyright (c) 2025-2026 TrustReady <hello@trustready.io>.
# SPDX-License-Identifier: MIT

set -euo pipefail

# Read Lima user from cidata (values are unquoted so we cannot source the file
# directly — LIMA_CIDATA_COMMENT may contain spaces).
LIMA_CIDATA_USER="$(sed -n 's/^LIMA_CIDATA_USER=//p' /mnt/lima-cidata/lima.env)"
export LIMA_CIDATA_USER

export DEBIAN_FRONTEND=noninteractive

GO_VERSION="1.27.1"
NODE_MAJOR=24
NPM_VERSION="12.0.2"

GOW_VERSION="v0.0.0-20260225145757-ff0f6779ab4c"
MKCERT_VERSION="v1.4.4"

apt-get update -qq
apt-get install -y -qq \
  build-essential \
  git \
  curl \
  jq \
  parallel \
  ca-certificates \
  gnupg \
  lsb-release \
  postgresql-client

if ! command -v docker &>/dev/null; then
  install -m 0755 -d /etc/apt/keyrings
  curl -fsSL https://download.docker.com/linux/ubuntu/gpg \
    | gpg --dearmor -o /etc/apt/keyrings/docker.gpg
  chmod a+r /etc/apt/keyrings/docker.gpg

  echo \
    "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] \
        https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" \
    | tee /etc/apt/sources.list.d/docker.list >/dev/null

  apt-get update -qq
  apt-get install -y -qq \
    docker-ce \
    docker-ce-cli \
    containerd.io \
    docker-buildx-plugin \
    docker-compose-plugin

  systemctl enable --now docker
fi

usermod -aG docker "${LIMA_CIDATA_USER:-lima}" 2>/dev/null || true

if [ ! -d "/usr/local/go" ] || ! /usr/local/go/bin/go version | grep -q "go${GO_VERSION}"; then
  rm -rf /usr/local/go
  ARCH=$(dpkg --print-architecture)
  curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-${ARCH}.tar.gz" \
    | tar -C /usr/local -xzf -
fi

cat >/etc/profile.d/go.sh <<'GOEOF'
export PATH="/usr/local/go/bin:$HOME/go/bin:$PATH"
GOEOF
chmod +x /etc/profile.d/go.sh

export PATH="/usr/local/go/bin:$PATH"
export HOME="${HOME:-/root}"

GOBIN=/usr/local/bin /usr/local/go/bin/go install "github.com/mitranim/gow@${GOW_VERSION}"

if ! command -v node &>/dev/null || ! node --version | grep -q "v${NODE_MAJOR}"; then
  curl -fsSL "https://deb.nodesource.com/setup_${NODE_MAJOR}.x" | bash -
  apt-get install -y -qq nodejs
fi

npm install -g "npm@${NPM_VERSION}"

if ! command -v mkcert &>/dev/null; then
  GOBIN=/usr/local/bin /usr/local/go/bin/go install "filippo.io/mkcert@${MKCERT_VERSION}"
fi
mkcert -install 2>/dev/null || true

LIMA_USER="${LIMA_CIDATA_USER:-lima}"
LIMA_HOME=$(eval echo "~${LIMA_USER}")
mkdir -p /root/.parallel "${LIMA_HOME}/.parallel"
touch /root/.parallel/will-cite "${LIMA_HOME}/.parallel/will-cite"
chown -R "${LIMA_USER}:${LIMA_USER}" "${LIMA_HOME}/.parallel"

# Generate sandbox-specific trustreadyd config and frontend .env files
VM_IP=$(ip -4 -j addr show dev lima0 | jq -r '.[0].addr_info[0].local')

su - "${LIMA_USER}" -c "export PATH=/usr/local/go/bin:\$HOME/go/bin:\$PATH && cd /workspace && make bin/trustreadyd-bootstrap"

make -C /workspace compose/step-ca/certs/root_ca.crt compose/keycloak/trustready-realm.json

mkdir -p /etc/trustreadyd

OAUTH2_SIGNING_KEY_PATH=/etc/trustreadyd/oauth2-signing-key.pem
if [ ! -f "${OAUTH2_SIGNING_KEY_PATH}" ]; then
  openssl genrsa -out "${OAUTH2_SIGNING_KEY_PATH}" 2048
  chmod 600 "${OAUTH2_SIGNING_KEY_PATH}"
fi

IDENTITY_FEDERATION_SIGNING_KEY_PATH=/etc/trustreadyd/identity-federation-signing-key.pem
if [ ! -f "${IDENTITY_FEDERATION_SIGNING_KEY_PATH}" ]; then
  openssl genrsa -out "${IDENTITY_FEDERATION_SIGNING_KEY_PATH}" 2048
  chmod 600 "${IDENTITY_FEDERATION_SIGNING_KEY_PATH}"
fi

# Load developer-specific overrides (not committed to repo).
if [ -f /workspace/.sandbox.env ]; then
  set -a
  # shellcheck source=/dev/null
  . /workspace/.sandbox.env
  set +a
fi

TRUSTREADYD_BASE_URL="http://${VM_IP}:8080" \
  TRUSTREADYD_AUTH_COOKIE_DOMAIN="${VM_IP}" \
  TRUSTREADYD_AUTH_COOKIE_SECURE=false \
  TRUSTREADYD_AUTH_COOKIE_SECRET="this-is-a-secure-secret-for-cookie-signing-at-least-32-bytes" \
  TRUSTREADYD_AUTH_PASSWORD_PEPPER="this-is-a-secure-pepper-for-password-hashing-at-least-32-bytes" \
  TRUSTREADYD_ENCRYPTION_KEY="thisisnotasecretAAAAAAAAAAAAAAAAAAAAAAAAAAA=" \
  TRUSTREADYD_OAUTH2_SERVER_SIGNING_KEY="$(cat "${OAUTH2_SIGNING_KEY_PATH}")" \
  TRUSTREADYD_IDENTITY_FEDERATION_ENABLED=true \
  TRUSTREADYD_IDENTITY_FEDERATION_SIGNING_KEY="$(cat "${IDENTITY_FEDERATION_SIGNING_KEY_PATH}")" \
  TRUSTREADYD_API_CORS_ALLOWED_ORIGINS="http://${VM_IP}:8080,http://${VM_IP}:5173,http://${VM_IP}:5174,http://${VM_IP}:5175" \
  TRUSTREADYD_AWS_ENDPOINT="http://127.0.0.1:8333" \
  TRUSTREADYD_AWS_ACCESS_KEY_ID="trustreadyd" \
  TRUSTREADYD_AWS_SECRET_ACCESS_KEY="thisisnotasecret" \
  TRUSTREADYD_AWS_USE_PATH_STYLE=true \
  TRUSTREADYD_ACME_DIRECTORY="https://127.0.0.1:9000/acme/acme/directory" \
  TRUSTREADYD_ACME_EMAIL="admin@trustready.io" \
  TRUSTREADYD_ACME_KEY_TYPE="EC256" \
  TRUSTREADYD_ACME_ROOT_CA="$(cat /workspace/compose/step-ca/certs/root_ca.crt)" \
  /workspace/bin/trustreadyd-bootstrap -output /etc/trustreadyd/config.yml

# trustreadyd runs as ${LIMA_USER} but bootstrap writes config.yml as root with 0600
# because it contains secrets. Transfer ownership so trustreadyd can read it.
chown "${LIMA_USER}:${LIMA_USER}" /etc/trustreadyd/config.yml "${OAUTH2_SIGNING_KEY_PATH}" "${IDENTITY_FEDERATION_SIGNING_KEY_PATH}"

# Bind-mount VM-local node_modules over the shared workspace to avoid
# platform conflicts between macOS host and Linux VM native binaries.
cat >/etc/systemd/system/trustready-node-modules.service <<EOF
[Unit]
Description=Bind-mount VM-local node_modules over workspace
DefaultDependencies=no
Before=trustready-console.service trustready-compliance-portal.service trustready-employee-portal.service

[Service]
Type=oneshot
RemainAfterExit=yes
ExecStartPre=/bin/mkdir -p /var/lib/trustready/node_modules /workspace/node_modules
ExecStart=/bin/mount --bind /var/lib/trustready/node_modules /workspace/node_modules
ExecStartPost=/bin/chown ${LIMA_USER}:${LIMA_USER} /var/lib/trustready/node_modules

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable --now trustready-node-modules.service

# Populate VM-local node_modules with Linux-native binaries (esbuild, etc.).
# The host's node_modules is macOS; `trustready-node-modules.service` bind-mounts an
# empty tree over /workspace/node_modules, and we install into it once here so
# the dev servers can run without cross-platform mismatches.
su - "${LIMA_USER}" -c "cd /workspace && npm ci"

# Generate go and ts files for trustreadyd and apps, and create embedded files for trustreadyd
make -C /workspace generate WITH_APPS=1
make -C /workspace embed

echo "VITE_API_URL=http://${VM_IP}:8080" >/workspace/apps/console/.env
echo "VITE_API_URL=http://${VM_IP}:8080" >/workspace/apps/compliance-portal/.env
echo "VITE_API_URL=http://${VM_IP}:8080" >/workspace/apps/employee-portal/.env

# Install systemd services for the sandbox
cat >/etc/systemd/system/trustready-stack.service <<EOF
[Unit]
Description=TrustReady Docker Compose Stack
Requires=docker.service
After=docker.service

[Service]
Type=oneshot
RemainAfterExit=yes
User=root
WorkingDirectory=/workspace
ExecStart=/usr/bin/docker compose -f compose.yaml up -d --wait
ExecStop=/usr/bin/docker compose -f compose.yaml down
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
EOF

cat >/etc/systemd/system/trustreadyd.service <<EOF
[Unit]
Description=TrustReady API Server
Requires=trustready-stack.service
After=trustready-stack.service

[Service]
Type=simple
User=${LIMA_USER}
WorkingDirectory=/workspace
ExecStartPre=/bin/bash -c 'for i in \$(seq 1 10); do pg_isready -h localhost -p 5432 -U trustreadyd -d trustreadyd -q && exit 0; sleep 1; done; echo "PostgreSQL not ready after 10s"; exit 1'
ExecStart=/usr/local/bin/gow -r=false run ./cmd/trustreadyd -cfg-file /etc/trustreadyd/config.yml -format pretty
Restart=on-failure
RestartSec=3s
Environment=PATH=/usr/local/go/bin:/usr/local/bin:/usr/bin:/bin

[Install]
WantedBy=multi-user.target
EOF

cat >/etc/systemd/system/trustready-console.service <<EOF
[Unit]
Description=TrustReady Console Dev Server
Requires=trustready-node-modules.service
After=trustready-node-modules.service trustreadyd.service

[Service]
Type=simple
User=${LIMA_USER}
WorkingDirectory=/workspace
ExecStart=/usr/bin/npm --workspace @trustready/console run dev -- --host 0.0.0.0
Restart=on-failure
RestartSec=3s

[Install]
WantedBy=multi-user.target
EOF

cat >/etc/systemd/system/trustready-compliance-portal.service <<EOF
[Unit]
Description=TrustReady Compliance Portal Dev Server
Requires=trustready-node-modules.service
After=trustready-node-modules.service trustreadyd.service

[Service]
Type=simple
User=${LIMA_USER}
WorkingDirectory=/workspace
ExecStart=/usr/bin/npm --workspace @trustready/compliance-portal run dev -- --host 0.0.0.0
Restart=on-failure
RestartSec=3s

[Install]
WantedBy=multi-user.target
EOF

cat >/etc/systemd/system/trustready-employee-portal.service <<EOF
[Unit]
Description=TrustReady Employee Portal Dev Server
Requires=trustready-node-modules.service
After=trustready-node-modules.service trustreadyd.service

[Service]
Type=simple
User=${LIMA_USER}
WorkingDirectory=/workspace
ExecStart=/usr/bin/npm --workspace @trustready/employee-portal run dev -- --host 0.0.0.0
Restart=on-failure
RestartSec=3s

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable --now trustready-stack.service
systemctl enable --now trustreadyd.service trustready-console.service trustready-compliance-portal.service trustready-employee-portal.service
