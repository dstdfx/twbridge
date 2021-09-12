#!/usr/bin/env bash

echo "==> Running unit tests..."
GO111MODULE=on go test -mod=mod -timeout=5m -race -v --count=1 ./...
if [[ $? -ne 0 ]]; then
    echo ""
    echo "Unit tests failed."
    exit 1
fi

exit 0
