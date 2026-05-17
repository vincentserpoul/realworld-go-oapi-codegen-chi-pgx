###########
# BUILDER #
###########

FROM --platform=${BUILDPLATFORM} golang:1.24.2 AS builder

# Install cross-compilation tools only if building for ARM architecture
RUN apt-get update && apt-get install -y \
    gcc-aarch64-linux-gnu libc6-dev-arm64-cross

WORKDIR /src


# Copy only go.mod and go.sum first
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .

ARG TARGETOS TARGETARCH
ARG BUILD_TIME LAST_MAIN_COMMIT_HASH

ENV CGO_ENABLED=0

ENV FLAGS="-s -w -X main.BuildTime=${BUILD_TIME} -X main.CommitHash=${LAST_MAIN_COMMIT_HASH}"

ARG BINARIES="realworld"

# Build all binaries for the specified architecture
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    export CGO_ENABLED=$CGO_ENABLED \
    export GOOS=$TARGETOS \
    export GOARCH=$TARGETARCH \
    export CC=$(if [ "${TARGETARCH}" = "arm64" ]; then echo aarch64-linux-gnu-gcc; else echo gcc; fi) && \
    mkdir /app && \
    for BINARY_NAME in ${BINARIES}; do \
        echo "Building ${BINARY_NAME} for OS: $TARGETOS, Architecture: $TARGETARCH" && \
        go build -installsuffix "static" -ldflags "${FLAGS}" -buildvcs=true -o /app/${BINARY_NAME}_test_network /src/cmd/${BINARY_NAME}/*.go; \
    done

#########
# FINAL #
#########

FROM gcr.io/distroless/cc-debian12:nonroot

ARG BINARY_NAME

COPY --chown=nonroot:nonroot ./config/${BINARY_NAME} /home/nonroot/config/${BINARY_NAME}

COPY --chown=nonroot:nonroot --from=builder /app /home/nonroot/app

USER nonroot

CMD ["/home/nonroot/app"]
