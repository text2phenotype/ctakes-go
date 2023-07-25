#!/usr/bin/env bash

# Source the bash.utils
source "${APP_BASE_DIR}/bin/bash.utils"

log.info "> Running unit tests..."
if [[ -e "fdl/test/unit" ]]; then
  python setup.py test -s fdl.tests.unit
  exit_code=$?
else
  log.info "> No tests exist at 'fdl/test/unit', skipping!"
  exit_code=0
fi
log.info "> Unit tests complete. Exited ('${exit_code}')"
exit ${exit_code}
