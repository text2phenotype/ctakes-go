#!/usr/bin/env bash

if [[ -x build-tools/bin/suite-runner ]]; then
    # export DOCKER_FROM_REPO='base-image'

    # Run the suite runner
    export DOCKER_NORMAL_BUILD='true'
    export DOCKER_BUILD_OPTIONS='--skip-from --target-repo fdl'
    build-tools/bin/suite-runner
else
    echo "The script does not exist or did not have execute permissions: build-tools/bin/suite-runner"
    stat build-tools/bin/suite-runner
fi
