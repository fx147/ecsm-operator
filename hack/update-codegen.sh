#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

# 项目根目录
SCRIPT_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
BOILERPLATE_PATH="${SCRIPT_ROOT}/hack/boilerplate.go.txt"

echo "暂无代码生成任务"