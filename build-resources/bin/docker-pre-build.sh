#!/usr/bin/env bash

set -e

# Source the bash.utils
source "${APP_BASE_DIR}/bin/bash.utils"

log.info "> Starting ${APP_NAME} build..."

log.info ">> Installing ${APP_NAME} dependencies..."

apt-get update -y

apt-get install -y \
  nginx

log.info ">> Creating application user..."
useradd -u 5001 -U -c "Text2Phenotype App User" -M -d $APP_BASE_DIR -s /bin/bash mdluser

# Add dumb-init to assist with proper signal handling
curl -Ls https://github.com/Yelp/dumb-init/releases/download/v1.2.0/dumb-init_1.2.0_amd64 -o /usr/local/bin/dumb-init
chmod +x /usr/local/bin/dumb-init

log.info ">> ${APP_NAME} dependency installation complete."

log.info ">> Cleaning up ${APP_NAME} build..."

log.info ">>> apt-get autoremove"
apt-get autoremove -y

log.info ">>> apt-get clean"
apt-get clean -y

log.info ">>> removing uncleaned apt files in '/var/lib/apt/lists/'"
rm -rf "/var/lib/apt/lists/*.*"
rm -rf "/var/lib/apt/lists/*"

log.info ">> ${APP_NAME} build cleanup complete."
