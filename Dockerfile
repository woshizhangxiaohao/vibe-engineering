# Railway Docker Build - yt-dlp transcript support
# v11 - root level Dockerfile
FROM golang:1.24 AS build

WORKDIR /go/src/vibe-backend

# Copy backend files from project root
COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ .
RUN CGO_ENABLED=0 GO111MODULE=on go build -o /go/bin/server ./cmd/server/main.go

# Runtime stage with yt-dlp
FROM debian:bookworm-slim

# Install yt-dlp and dependencies for YouTube transcript extraction
RUN apt-get update && apt-get install -y \
    python3 \
    python3-pip \
    ffmpeg \
    ca-certificates \
    && pip3 install --break-system-packages --no-cache-dir yt-dlp \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

COPY --from=build /go/bin/server /server
COPY --from=build /go/src/vibe-backend/migrations /migrations
COPY entrypoint.sh /entrypoint.sh

# Ensure executables have correct permissions
RUN chmod +x /server /entrypoint.sh

EXPOSE 8080

# Use shell script entrypoint for better logging
ENTRYPOINT ["/entrypoint.sh"]
