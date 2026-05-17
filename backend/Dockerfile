# Recipes API image.
#
# Build context MUST be the backend/ directory so the Go workspace
# (go.work) and all four modules (api, cli, repo, types) are available.
#
# Example:
#   docker build -t recipes-api ./backend
#   docker run --rm -p 4000:4000 \
#     -e DB_HOST=recipes-db -e DB_PORT=5432 \
#     -e DB_USER=postgres  -e DB_PASSWORD=postgres \
#     -e DB_NAME=recipes \
#     recipes-api
#
# In Kubernetes, point DB_HOST at the StatefulSet's headless/ClusterIP
# Service (e.g. DB_HOST=recipes-db). The API itself is stateless and
# can run as a Deployment.

# ---------- build stage ----------
FROM golang:1.26-alpine AS build

WORKDIR /src

# Prime the module cache. Copying go.{work,mod,sum} files first lets
# Docker reuse the download layer when only source files change.
COPY go.work go.work.sum ./
COPY api/go.mod   api/go.sum   api/
COPY cli/go.mod   cli/go.sum   cli/
COPY repo/go.mod  repo/go.sum  repo/
COPY types/go.mod              types/

RUN go mod download

COPY api/   api/
COPY repo/  repo/
COPY types/ types/

# Static, stripped binary — works in lightweight shell-capable runtimes.
ENV CGO_ENABLED=0 GOOS=linux
RUN go build -trimpath -ldflags="-s -w" -o /out/recipes-api ./api

# ---------- runtime stage ----------
FROM alpine:3.20

# Bind on all interfaces so the container is reachable via the k8s Service.
ENV API_ADDR=:4000

# DB connection. DB_HOST should point at the Postgres StatefulSet's
# (headless) Service — e.g. DB_HOST=recipes-db. Inject DB_USER and
# DB_PASSWORD from a k8s Secret; never bake credentials into the image.
ENV DB_HOST=recipes-db \
    DB_PORT=5432 \
    DB_NAME=recipes

EXPOSE 4000

COPY --from=build /out/recipes-api /usr/local/bin/recipes-api

RUN addgroup -S -g 65532 nonroot \
    && adduser -S -D -H -u 65532 -G nonroot -h /app -s /bin/sh nonroot
USER nonroot:nonroot
ENTRYPOINT ["/usr/local/bin/recipes-api"]
