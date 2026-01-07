FROM golang:1.24 AS build

WORKDIR /go/src/vibe-backend

# NO-OP UNLESS DEPLOYMENT_FIX: This code works on Railway. touch ONLY if it breaks Railway production build/start.
COPY backend/ .

RUN --mount=type=cache,id=s/3a2cccef-6f6d-4cb8-aaaf-a49ddec22628-/go/pkg/mod,target=/go/pkg/mod \
    go mod download

# https://docs.railway.com/guides/dockerfiles#cache-mounts
RUN --mount=type=cache,id=s/3a2cccef-6f6d-4cb8-aaaf-a49ddec22628-/go/pkg/mod,target=/go/pkg/mod \
    --mount=type=cache,id=s/3a2cccef-6f6d-4cb8-aaaf-a49ddec22628-/root/.cache/go-build,target=/root/.cache/go-build \
    CGO_ENABLED=0 GO111MODULE=on go build -o /go/bin/server ./cmd/server/main.go
# NO-OP UNLESS DEPLOYMENT_FIX: This code works on Railway. touch ONLY if it breaks Railway production build/start.

FROM gcr.io/distroless/static-debian13

COPY --from=build /go/bin/server /
COPY --from=build /go/src/vibe-backend/migrations /migrations

EXPOSE 8080

CMD ["/server"]
