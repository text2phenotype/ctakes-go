### Install go lang reqs into a temp layer
# FROM golang:buster as builder
# ENV APP_BASE_DIR="/app"

# COPY --chown=5001:5001 ./src/ "${APP_BASE_DIR}"/src/

# RUN go get
# RUN chown -R 5001:5001 /go


###  Start actual build
# Use a reasonable FROM image
ARG IMAGE_FROM_TAG
FROM golang:buster

# Create a list of build arguments
ARG APP_ENVIRONMENT
ARG APP_GIT_SHA
ARG APP_IS_DEBUG
ARG APP_SERVICE
ARG IMAGE_FROM_TAG

# Set environment variables
# UNIVERSE_IS_VERBOSE enables log level INFO.
ENV UNIVERSE_IS_VERBOSE=true

### Application metadata
ENV APP_ENVIRONMENT="${APP_ENVIRONMENT:-prod}"
ENV APP_GIT_SHA="${APP_GIT_SHA:-unset}"
ENV APP_IS_DEBUG="${APP_IS_DEBUG:-False}"
ENV APP_NAME="FDL"

# APP_SERVICE
ENV APP_SERVICE="${APP_SERVICE:-FDL}"

### File path locations
ENV APP_BASE_DIR="/app"
ENV MDL_COMN_DATA_ROOT="/tmp"
ENV PATH="${APP_BASE_DIR}/bin/:${PATH}"

### App vars required to run
ENV FDL_DICTIONARY_PATH="$APP_BASE_DIR/resources/dictionaries"
ENV FDL_CONFIG_PATH="$APP_BASE_DIR/config"
ENV MDL_COMN_LOGLEVEL=INFO

# Set some container options
WORKDIR "${APP_BASE_DIR}"
RUN chown 5001:5001 "${APP_BASE_DIR}"

###  This section should be ordered in such a way that the least likely 
### operation to change should be first.

# Copy large resource files first
COPY --chown=5001:5001 ./resources "${APP_BASE_DIR}"/resources/

# Copy required build files early
COPY --chown=5001:5001 ./build-resources/bin/ "${APP_BASE_DIR}"/bin/
COPY --chown=5001:5001 ./build-tools/bin/ "${APP_BASE_DIR}"/bin/

# Run the pre-build to install packages that rarely change.
RUN "${APP_BASE_DIR}/bin/docker-pre-build.sh"

# Copy the application build directory previously built
# COPY --from=builder /go /go

# Copy the application code, should be ordered in least likely to change first
COPY --chown=5001:5001 ./bin/ "${APP_BASE_DIR}"/bin/
COPY --chown=5001:5001 ./config "${APP_BASE_DIR}"/config/
COPY --chown=5001:5001 ./src/ "${APP_BASE_DIR}"/src/

WORKDIR "${APP_BASE_DIR}/src"
RUN go get && chown -R 5001:5001 /go

USER 5001

# dumb-init is used to assist with proper signal handling, without
# it we will not kill the other processes
ENTRYPOINT ["/usr/local/bin/dumb-init","--"]

WORKDIR "${APP_BASE_DIR}/src"

# This command is what launches the service by default.
CMD ["/bin/bash", "-c", "go run ./."]
