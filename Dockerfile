###########
# BUILDER #
###########

FROM --platform=${BUILDPLATFORM} golang:1.21.4 AS builder

WORKDIR /src

COPY . .

ARG BINARY_NAME
ARG TARGETOS TARGETARCH
ARG LAST_MAIN_COMMIT_HASH LAST_MAIN_COMMIT_TIME
ARG GLOBAL_VAR_PKG

ENV FLAG="-X ${GLOBAL_VAR_PKG}.CommitTime=${LAST_MAIN_COMMIT_TIME}"
ENV FLAG="$FLAG -X ${GLOBAL_VAR_PKG}.CommitHash=${LAST_MAIN_COMMIT_HASH}"

RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build \
    -installsuffix 'static' \
    -ldflags "-s -w $FLAG" \
    -buildvcs=true \
    -o /app ./cmd/${BINARY_NAME}/*.go

#########
# FINAL #
#########

FROM gcr.io/distroless/cc-debian12:nonroot

ARG BINARY_NAME

COPY ./config/${BINARY_NAME} /config/${BINARY_NAME}

COPY --from=builder /app /app

USER nonroot

CMD ["/app"]
