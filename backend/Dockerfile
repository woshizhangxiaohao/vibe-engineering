# Railway Docker Build - yt-dlp transcript support
# FORCE REBUILD: 2026-01-13 14:10:00 UTC - v9 fix path
FROM golang:1.24 AS build

# IMPORTANT: Rebuild required for yt-dlp support - 2026-01-13-v9
ARG BUILD_VERSION=2026-01-13-v9-fix-copy-path
ARG CACHEBUST=20260113-141000

WORKDIR /go/src/vibe-backend

# Copy from current directory (works when build context is backend/)
COPY . .

RUN --mount=type=cache,id=s/3a2cccef-6f6d-4cb8-aaaf-a49ddec22628-/go/pkg/mod,target=/go/pkg/mod \
    go mod download

# https://docs.railway.com/guides/dockerfiles#cache-mounts
RUN --mount=type=cache,id=s/3a2cccef-6f6d-4cb8-aaaf-a49ddec22628-/go/pkg/mod,target=/go/pkg/mod \
    --mount=type=cache,id=s/3a2cccef-6f6d-4cb8-aaaf-a49ddec22628-/root/.cache/go-build,target=/root/.cache/go-build \
    CGO_ENABLED=0 GO111MODULE=on go build -o /go/bin/server ./cmd/server/main.go

FROM debian:bookworm-slim

# Build version for cache busting
ARG BUILD_VERSION=2026-01-13-v9-fix-copy-path
LABEL build.version="${BUILD_VERSION}"

# Install yt-dlp and dependencies for YouTube transcript extraction
RUN apt-get update && apt-get install -y \
    python3 \
    python3-pip \
    ffmpeg \
    ca-certificates \
    && pip3 install --break-system-packages --no-cache-dir yt-dlp \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

COPY --from=build /go/bin/server /
COPY --from=build /go/src/vibe-backend/migrations /migrations

EXPOSE 8080

CMD ["/server"]
