FROM golang:1.21 AS build

WORKDIR /go/src/vibe-backend

# Copy go mod files first for better caching
COPY backend/go.mod backend/go.sum ./

# Download dependencies (fresh cache v3)
RUN --mount=type=cache,id=s/3a2cccef-6f6d-4cb8-aaaf-a49ddec22628-v3/go/pkg/mod,target=/go/pkg/mod \
    go mod download

# Force rebuild: 2026-01-12-19:08-v3
RUN echo "Build timestamp: 2026-01-12-19:08-v3 - OAuth callback route included"

# Copy source code
COPY backend/ .

# https://docs.railway.com/guides/dockerfiles#cache-mounts
# Using v3 cache ID to force complete rebuild
RUN --mount=type=cache,id=s/3a2cccef-6f6d-4cb8-aaaf-a49ddec22628-v3/go/pkg/mod,target=/go/pkg/mod \
    --mount=type=cache,id=s/3a2cccef-6f6d-4cb8-aaaf-a49ddec22628-v3/root/.cache/go-build,target=/root/.cache/go-build \
    CGO_ENABLED=0 GO111MODULE=on go build -mod=mod -o /go/bin/server ./cmd/server/main.go
# NO-OP UNLESS DEPLOYMENT_FIX: This code works on Railway. touch ONLY if it breaks Railway production build/start.

FROM gcr.io/distroless/static-debian13

COPY --from=build /go/bin/server /
COPY --from=build /go/src/vibe-backend/migrations /migrations

EXPOSE 8080

CMD ["/server"]
