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

# Run stage: only the binary, no shell or OS packages.
FROM scratch
COPY --from=build /hangman /hangman
USER 65534:65534
EXPOSE 8080
ENTRYPOINT ["/hangman"]
