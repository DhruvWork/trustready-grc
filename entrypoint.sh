#!/bin/bash
set -e

# Configuration file path
CONFIG_FILE="${CONFIG_FILE:-/etc/trustreadyd/config.yml}"

# When TRUSTREADYD_ENCRYPTION_KEY is set, always (re)generate the config from env vars.
# This includes literal values and aws:// / awssm:// / awsps:// AWS references.
# trustreadyd-bootstrap reads every TRUSTREADYD_* var; the entrypoint only checks this one
# to decide whether to run it, including when a stale config file exists on a
# persistent volume. When it is unset, fall back to an existing config file
# (e.g., mounted from a ConfigMap).
if [ -n "$TRUSTREADYD_ENCRYPTION_KEY" ]; then
  echo "Generating configuration file from environment variables at: $CONFIG_FILE"
  trustreadyd-bootstrap -output "$CONFIG_FILE"
elif [ -f "$CONFIG_FILE" ]; then
  echo "Using existing configuration file at: $CONFIG_FILE"
else
  echo "Error: TRUSTREADYD_ENCRYPTION_KEY is unset and no config file found at $CONFIG_FILE" >&2
  exit 1
fi

# Execute trustreadyd with the generated config
exec trustreadyd -cfg-file "$CONFIG_FILE" "$@"
