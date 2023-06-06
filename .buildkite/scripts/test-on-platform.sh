#!/bin/bash

set -euo pipefail

go version
echo "go-${GO_VERSION}"
make lint
make unit
