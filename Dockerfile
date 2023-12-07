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

RUN --mount=type=cache,target=/root/.cache/go-build \
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

COPY ./config/${BINARY_NAME} /config/${BINARY_NAME}

COPY --from=builder /app /app

USER nonroot

CMD ["/app"]
