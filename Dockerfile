FROM --platform=$BUILDPLATFORM tonistiigi/xx:1.4.0 AS xx

# Build the manager binary
FROM golang:1.22-alpine as builder
SHELL ["/bin/ash", "-o", "pipefail", "-c"]
COPY --from=xx / /
ARG TARGETPLATFORM
ARG COMPRESS_BINARY

RUN apk add --update --no-cache upx

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

# Copy the go source
COPY cmd/main.go cmd/main.go
COPY api/ api/
COPY internal/ internal/

# Build
# the GOARCH has not a default value to allow the binary be built according to the host where the command
# was called. For example, if we call make docker-build in a local env which has the Apple Silicon M1 SO
# the docker BUILDPLATFORM arg will be linux/arm64 when for Apple x86 it will be linux/amd64. Therefore,
# by leaving it empty we can ensure that the container and binary shipped on it will have the same platform.
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg \
    CGO_ENABLED=0 xx-go build -trimpath -ldflags '-w -s' -o manager-uncompressed cmd/main.go
RUN xx-verify --static manager-uncompressed
RUN if [ -n "$COMPRESS_BINARY" ]; then \
      upx --best --lzma -o manager manager-uncompressed; \
    else \
      mv manager-uncompressed manager; \
    fi

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
COPY --from=builder /workspace/manager /usr/local/bin/cloudflared-operator
USER 65532:65532

ENTRYPOINT ["cloudflared-operator"]
