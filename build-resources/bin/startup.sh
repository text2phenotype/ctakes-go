#!/usr/bin/env bash

# Source the bash.utils
source "${APP_BASE_DIR}/bin/bash.utils"

log.info "> Starting ${APP_NAME}..."

export PATH="${APP_BASE_DIR}/.local/bin:$PATH"

export GUNICORN_SOCKET_PATH="unix:${APP_BASE_DIR}/fdl.sock"

# Use "gthread" for sync flask App
#export GUNICORN_WORKER_CLASS="${GUNICORN_WORKER_CLASS:-gthread}"

# Use "aiohttp.worker.GunicornWebWorker" for async AioHttpApp
export GUNICORN_WORKER_CLASS="${GUNICORN_WORKER_CLASS:-aiohttp.worker.GunicornWebWorker}"

# Default threads to 1 or it forces gthreads as the worker class
export GUNICORN_WORKER_THREADS="${GUNICORN_WORKER_THREADS:-100}"
export GUNICORN_WORKER_TIMEOUT="${GUNICORN_WORKER_TIMEOUT:-600}"
export GUNICORN_WORKERS="${GUNICORN_WORKERS:-1}"

# Port Nginx will listen on
export NGINX_LISTEN_PORT="${NGINX_LISTEN_PORT:-8080}"

declare nginx_default_conf="/etc/nginx/sites-enabled/default"

log.info ">>> Installing nginx proxy config..."
log.info "GUNICORN_SOCKET_PATH: ${GUNICORN_SOCKET_PATH}"

j2 "${APP_BASE_DIR}/build-resources/config/nginx-proxy.conf.j2" > "${nginx_default_conf}"
log.info ">> nginx configuration complete."

log.info ">> Starting nginx..."
/usr/sbin/nginx

log.info ">> Launching guincorn workers..."
log.info "Worker Class  : ${GUNICORN_WORKER_CLASS}"
log.info "Worker Timeout: ${GUNICORN_WORKER_TIMEOUT}"
log.info "Worker Number : ${GUNICORN_WORKERS}"
if [[ "$GUNICORN_WORKER_CLASS" == "gthread" ]]; then
  log.info "Worker Threads : ${GUNICORN_WORKER_THREADS}"
fi

time gunicorn \
  --timeout ${GUNICORN_WORKER_TIMEOUT} \
  --workers ${GUNICORN_WORKERS} \
  --worker-class ${GUNICORN_WORKER_CLASS} \
  --threads ${GUNICORN_WORKER_THREADS} \
  --bind ${GUNICORN_SOCKET_PATH} 'fdl.__main__:create_app().app'

log.warn "> ${APP_NAME} stopped!"
