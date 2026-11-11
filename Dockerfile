# syntax=docker/dockerfile:1.7
# Two images from one file (ADR-0027), picked with --target:
#   web  nginx serving the built SPA and passing /api to the API container
#   api  the Go server on a minimal non-root runtime (the default target)
# `make up` runs both with compose (compose.yaml).
#
#   docker build --target web -t middy-web .
#   docker build --target api -t middy-api .

ARG NODE_VERSION=22
ARG GO_VERSION=1.27

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
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -trimpath \
      -ldflags "-s -w -X github.com/hassaku63/middy/backend/internal/httpapi.Version=${VERSION}" \
      -o /out/server ./cmd/server

# --- web: nginx and the SPA (ADR-0027) --------------------------------------
# The image renders deploy/nginx/templates into conf.d at start from PORT and
# API_UPSTREAM (a URL, e.g. http://api:8081) and RESOLVER (the DNS that
# resolves it). It runs as a non-root user.
FROM nginxinc/nginx-unprivileged:1.31-alpine AS web
COPY --from=frontend /src/frontend/dist /usr/share/nginx/html
COPY deploy/nginx/templates/ /etc/nginx/templates/
COPY deploy/nginx/snippets/ /etc/nginx/snippets/
ENV PORT=8080 RESOLVER=127.0.0.11
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --start-interval=1s --retries=3 \
  CMD ["sh", "-c", "wget -q -O /dev/null \"http://127.0.0.1:${PORT}/healthz\" || exit 1"]

# --- api: the Go server ------------------------------------------------------
FROM gcr.io/distroless/static-debian12:nonroot AS api
WORKDIR /app
COPY --from=backend /out/server /app/server
ENV PORT=8080 LLM_PROVIDER=fake
EXPOSE 8080
# The image's nonroot user, by number: a platform that checks for a
# non-root user (Kubernetes runAsNonRoot) can only verify a numeric one.
USER 65532:65532
# distroless has no shell or curl; the server exposes a probe flag instead.
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --start-interval=1s --retries=3 \
  CMD ["/app/server", "-healthcheck"]
ENTRYPOINT ["/app/server"]
