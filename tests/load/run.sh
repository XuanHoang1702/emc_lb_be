#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

if [ -z "$1" ]; then
    echo "Usage: ./run.sh scripts/<script_name>.js"
    exit 1
fi

SCRIPT_PATH="$1"

if [ ! -f "$SCRIPT_PATH" ]; then
    echo "Error: script $SCRIPT_PATH not found"
    exit 1
fi

echo "Running k6 test: $SCRIPT_PATH"
# Run k6 in docker using host network to access localhost NGINX
docker run --rm -i --network host -v "${SCRIPT_DIR}:/load" -e BASE_URL="http://localhost" grafana/k6 run "/load/$SCRIPT_PATH"
