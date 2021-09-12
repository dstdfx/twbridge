#!/usr/bin/env bash

echo "==> Building twbridge binary..."
GO111MODULE=on CGO_ENABLED=0 \
go build -mod=mod -a -installsuffix cgo -ldflags \
    "-X github.com/dstdfx/twbridge/cmd/twbridge/app.buildGitCommit=$(git rev-parse HEAD) \
    -X github.com/dstdfx/twbridge/cmd/twbridge/app.buildGitTag=$(git describe --abbrev=0) \
    -X github.com/dstdfx/twbridge/cmd/twbridge/app.buildDate=$(date +%Y%m%d)" \
    -o twbridge ./cmd/twbridge
