#!/bin/bash
set -euo pipefail
cd "$(dirname "$0")/.."

OAPI="$(go env GOPATH)/bin/oapi-codegen"
if [ ! -x "$OAPI" ]; then
  echo "oapi-codegen not found at $OAPI"
  exit 1
fi

mkdir -p internal/generated/openapi

"$OAPI" -generate types,chi-server -package openapi \
  openapi/openapi-draft.yaml \
  > internal/generated/openapi/types.gen.go

echo "✅ Codegen complete: internal/generated/openapi/types.gen.go"
