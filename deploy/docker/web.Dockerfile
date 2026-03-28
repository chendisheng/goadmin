# syntax=docker/dockerfile:1.7

FROM node:20-alpine AS builder

WORKDIR /src/web

ARG VITE_APP_TITLE=GoAdmin
ARG VITE_API_BASE_URL=/api/v1

ENV VITE_APP_TITLE=${VITE_APP_TITLE} \
    VITE_API_BASE_URL=${VITE_API_BASE_URL}

COPY web/package*.json ./
RUN if [ -f package-lock.json ]; then npm ci --no-audit --no-fund; else npm install --no-audit --no-fund; fi

COPY web/ ./
RUN npm run build

FROM nginx:1.27-alpine AS runtime

RUN apk add --no-cache curl

ENV API_UPSTREAM=http://backend:8080

COPY deploy/docker/web-nginx.conf /etc/nginx/templates/default.conf.template
COPY --from=builder /src/web/dist /usr/share/nginx/html

EXPOSE 80

HEALTHCHECK --interval=10s --timeout=3s --retries=5 --start-period=10s CMD ["curl", "-fsS", "http://127.0.0.1/healthz"]

CMD ["nginx", "-g", "daemon off;"]
