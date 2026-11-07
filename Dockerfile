# syntax=docker/dockerfile:1.7
# Single-binary image (ADR-0008): build the frontend, embed it into the Go
# server, ship only the binary on a minimal non-root runtime.
#
#   docker build -t middy .
#   docker run --rm -p 8080:8080 middy

ARG NODE_VERSION=22
ARG GO_VERSION=1.24

# --- frontend ---------------------------------------------------------------
FROM node:${NODE_VERSION}-alpine AS frontend
WORKDIR /src/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN --mount=type=cache,target=/root/.npm npm ci --no-audit --no-fund
COPY frontend/ ./
RUN npm run build

# --- backend ----------------------------------------------------------------
FROM golang:${GO_VERSION}-alpine AS backend
ARG VERSION=dev
WORKDIR /src/backend
COPY backend/go.mod backend/go.su[m] ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY backend/ ./
COPY --from=frontend /src/frontend/dist ./internal/web/dist
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -trimpath \
      -ldflags "-s -w -X github.com/hassaku63/middy/backend/internal/httpapi.Version=${VERSION}" \
      -o /out/server ./cmd/server

# --- runtime ----------------------------------------------------------------
FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=backend /out/server /app/server
ENV PORT=8080 LLM_PROVIDER=fake
EXPOSE 8080
USER nonroot:nonroot
# distroless has no shell or curl; the server exposes a probe flag instead.
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD ["/app/server", "-healthcheck"]
ENTRYPOINT ["/app/server"]
