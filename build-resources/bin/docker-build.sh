#!/usr/bin/env bash

### THIS FILE NO LONGER USED

set -e

APPUSER=$(id -un 5001)

# Source the bash.utils
source "${APP_BASE_DIR}/bin/bash.utils"

# log.info ">> Chown'ing ${APP_BASE_DIR}..."
# chown -R $APPUSER:$APPUSER "${APP_BASE_DIR}"
# log.info ">> Done Chown'ing ${APP_BASE_DIR}..."

log.info ">> Temp linter install..."
go get github.com/golangci/golangci-lint/cmd/golangci-lint

log.info ">> Installing application..."
cd "$APP_BASE_DIR/src"
go get

log.info ">> Chown'ing go files to $APPUSER ..."
chown -R $APPUSER:$APPUSER "/go"
log.info ">> Done chown'ing go files to $APPUSER ..."

### NGINX - Not currently used.
# log.info ">> Setting up nginx config..."
# cp -v "${APP_BASE_DIR}/build-resources/config/nginx.conf" "/etc/nginx/nginx.conf"
# log.info ">> Done Setting up nginx config"

# log.info ">> Removing default nginx site..."
# rm -f "/etc/nginx/sites-available/default"
# rm -f "/etc/nginx/sites-enabled/default"
# log.info ">> Done removing default nginx site"

# log.info ">> Chown'ing nginx files to $APPUSER ..."
# touch /etc/nginx/sites-enabled/default
# chown -R $APPUSER:$APPUSER /etc/nginx/sites-enabled/default /var/log/nginx /var/lib/nginx
# log.info ">> Done chown'ing nginx files to $APPUSER ..."

log.info "> ${APP_NAME} installation complete."
