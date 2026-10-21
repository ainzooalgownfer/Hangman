# Build stage: compile a static binary. It runs on the build machine's own
# architecture and cross-compiles for the target, so multi-arch builds are fast.
FROM --platform=$BUILDPLATFORM golang:1.27-alpine AS build
ARG TARGETOS TARGETARCH
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -ldflags="-s -w" -o /hangman ./cmd/server

# Run stage: only the binary and its config files, no shell or OS packages.
# The server reads config.yaml and words.json from its working directory.
FROM scratch
WORKDIR /app
COPY --from=build /hangman /app/hangman
COPY config.yaml words.json /app/
USER 65534:65534
EXPOSE 8080
ENTRYPOINT ["/app/hangman"]
