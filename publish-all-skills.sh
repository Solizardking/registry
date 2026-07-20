#!/usr/bin/env bash
# Back-compat entrypoint: publish Skill Hub skills into agentregistry.
# Prefer: ./scripts/publish-skillhub-skills.sh
#
# Auto-detects skillhub at:
#   $SKILLHUB_DIR
#   ~/Downloads/skillhub-main
#   /Users/8bit/Downloads/skillhub-main
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
exec "$ROOT/scripts/publish-skillhub-skills.sh" "$@"
