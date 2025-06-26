#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

# 项目根目录
SCRIPT_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
BOILERPLATE_PATH="${SCRIPT_ROOT}/hack/boilerplate.go.txt"

DEEPCOPY_GEN="$(go env GOPATH)/bin/deepcopy-gen"

"${DEEPCOPY_GEN}" -i ${SCRIPT_ROOT}/pkg/apis/ecsm/v1 -O zz_generated.deepcopy --output-file-base zz_generated.deepcopy --output-package ./pkg/apis/ecsm/v1 --go-header-file hack/boilerplate.go.txt