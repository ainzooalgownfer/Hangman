# Build stage: compile a static binary.
FROM golang:1.27-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /hangman ./cmd/server

# Run stage: only the binary, no shell or OS packages.
FROM scratch
COPY --from=build /hangman /hangman
USER 65534:65534
EXPOSE 8080
ENTRYPOINT ["/hangman"]
