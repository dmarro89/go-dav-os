#!/bin/bash
# examples/run_docker.sh
# Build and run the OS using the Docker-based toolchain.
set -euo pipefail

# Build and run through Docker
echo "Building and running the OS inside Docker toolchain..."
make docker-run
