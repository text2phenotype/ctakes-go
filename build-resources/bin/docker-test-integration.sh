#!/usr/bin/env bash

# Source the bash.utils
source "${APP_BASE_DIR}/bin/bash.utils"

log.info "> Running integration tests..."

if [[ -e "fdl/test/integration" ]]; then
  python setup.py test -s fdl.tests.integration
  exit_code=$?
else
  log.info "> No tests exist at 'fdl/test/integration', skipping!"
  exit_code=0
fi

log.info "> Integration tests complete. Exited ('${exit_code}')"
exit ${exit_code}
