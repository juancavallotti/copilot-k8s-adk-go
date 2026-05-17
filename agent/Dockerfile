# Recipes agent image.
#
# Build context MUST be the repository root so this image can compile the
# top-level agent module and the existing backend/cli module.
#
# Example:
#   docker build -t recipes-agent -f agent/Dockerfile .

# ---------- build stage ----------
FROM golang:1.26-alpine AS build

WORKDIR /src

# Prime the agent module cache.
COPY agent/go.mod agent/go.sum agent/
WORKDIR /src/agent
RUN go mod download

# Prime the backend workspace cache for recipes-cli.
WORKDIR /src/backend
COPY backend/go.work backend/go.work.sum ./
COPY backend/api/go.mod  backend/api/go.sum  api/
COPY backend/cli/go.mod  backend/cli/go.sum  cli/
COPY backend/repo/go.mod backend/repo/go.sum repo/
COPY backend/types/go.mod              types/
RUN go mod download

# Build both static binaries.
WORKDIR /src
COPY agent/ agent/
COPY backend/cli/ backend/cli/
COPY backend/repo/ backend/repo/
COPY backend/types/ backend/types/

ENV CGO_ENABLED=0 GOOS=linux
WORKDIR /src/agent
RUN go build -trimpath -ldflags="-s -w" -o /out/recipes-agent .
WORKDIR /src/backend
RUN go build -trimpath -ldflags="-s -w" -o /out/recipes-cli ./cli

# ---------- runtime stage ----------
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

# Bind on all interfaces so the container is reachable via the k8s Service.
ENV AGENT_ADDR=:4100 \
    AGENT_MODEL=gemini-3.1-flash-lite \
    AGENT_IMAGE_MODEL=gemini-3.1-flash-image-preview \
    AGENT_INSTRUCTION_PATH=prompts/recipe_copilot.md

# DB connection used by recipes-cli when the agent invokes it.
ENV DB_HOST=recipes-db \
    DB_PORT=5432 \
    DB_NAME=recipes

EXPOSE 4100

COPY --from=build /out/recipes-agent /usr/local/bin/recipes-agent
COPY --from=build /out/recipes-cli /usr/local/bin/recipes-cli
COPY --from=build /src/agent/prompts/ /app/prompts/

USER nonroot:nonroot
ENTRYPOINT ["/usr/local/bin/recipes-agent"]
