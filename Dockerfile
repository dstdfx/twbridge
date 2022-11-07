FROM golang:1.19.3 AS build
WORKDIR /workspace
ENV GO111MODULE=on
ENV CGO_ENABLED=0
COPY . .
RUN go build -mod=mod -a -installsuffix cgo -ldflags \
        "-X github.com/dstdfx/twbridge/cmd/twbridge/app.buildGitCommit=$(git rev-parse HEAD) \
        -X github.com/dstdfx/twbridge/cmd/twbridge/app.buildGitTag=$(git describe --abbrev=0) \
        -X github.com/dstdfx/twbridge/cmd/twbridge/app.buildDate=$(date +%Y%m%d)" \
        -o twbridge ./cmd/twbridge

FROM alpine:3.15.2
RUN apk add --no-cache ca-certificates
COPY --from=build /workspace/twbridge /usr/local/bin/twbridge
ENTRYPOINT ["twbridge"]
