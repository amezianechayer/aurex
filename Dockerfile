# syntax=docker/dockerfile:1
# Multi-stage build: compile the corren binary from source, then ship a slim
# runtime. CGO is required (mattn/go-sqlite3); we build on a glibc (Debian) image
# because the bundled sqlite amalgamation fails to compile against musl/alpine
# with recent gcc. Runtime is debian-slim for ABI consistency.
FROM golang:1.21-bookworm AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o /out/corren .

FROM debian:bookworm-slim
RUN apt-get update \
    && apt-get install -y --no-install-recommends curl ca-certificates \
    && rm -rf /var/lib/apt/lists/*
COPY --from=build /out/corren /usr/local/bin/corren
EXPOSE 3068
ENTRYPOINT ["corren"]
CMD ["server", "start"]
