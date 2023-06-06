#!/bin/bash

set -euo pipefail

go version
make lint
make unit
