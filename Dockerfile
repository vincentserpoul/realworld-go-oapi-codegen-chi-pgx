###########
# BUILDER #
###########

FROM --platform=${BUILDPLATFORM} golang:1.21.4 AS builder

WORKDIR /src

COPY . .

ARG BINARY_NAME
ARG TARGETOS TARGETARCH
ARG BUILD_TIME LAST_MAIN_COMMIT_HASH

ENV FLAG="-X main.BuildTime=${BUILD_TIME}"
ENV FLAG="$FLAG -X main.CommitHash=${LAST_MAIN_COMMIT_HASH}"

ENV GOCACHE=/root/.cache/go-build

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build \
    -installsuffix 'static' \
    -ldflags "-s -w $FLAG" \
    -buildvcs=true \
    -o /app /src/cmd/${BINARY_NAME}/*.go

#########
# FINAL #
#########

FROM gcr.io/distroless/cc-debian12:nonroot

ARG BINARY_NAME

COPY --chown=nonroot:nonroot ./config/${BINARY_NAME} /home/nonroot/config/${BINARY_NAME}

COPY --chown=nonroot:nonroot --from=builder /app /home/nonroot/app

USER nonroot

CMD ["/home/nonroot/app"]
