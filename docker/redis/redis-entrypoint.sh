#!/usr/bin/env sh
set -eu

TEMPLATE_PATH="/usr/local/etc/redis/redis.conf.template"
CONFIG_PATH="/usr/local/etc/redis/redis.conf"

escape_sed_repl() {
  # Escape sed replacement special chars to safely inject passwords
  printf '%s' "$1" | sed -e 's/[\\&|]/\\&/g'
}

requirepass_line=""
masterauth_line=""

if [ "${REDIS_PASSWORD:-}" != "" ]; then
  requirepass_line="requirepass ${REDIS_PASSWORD}"
fi

# Allow overriding master auth separately (defaults to unset)
if [ "${REDIS_MASTER_PASSWORD:-}" != "" ]; then
  masterauth_line="masterauth ${REDIS_MASTER_PASSWORD}"
fi

requirepass_escaped=$(escape_sed_repl "${requirepass_line}")
masterauth_escaped=$(escape_sed_repl "${masterauth_line}")

sed \
  -e "s|__REDIS_REQUIREPASS__|${requirepass_escaped}|" \
  -e "s|__REDIS_MASTERAUTH__|${masterauth_escaped}|" \
  "${TEMPLATE_PATH}" > "${CONFIG_PATH}"

exec redis-server "${CONFIG_PATH}"
