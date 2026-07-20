# ===========================================
# Stage 1: Build frontend
# ===========================================
FROM node:22-alpine AS frontend-builder

WORKDIR /app/frontend

RUN npm install -g pnpm@latest

COPY frontend/package.json frontend/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile --prefer-offline

COPY frontend/ .
RUN pnpm run build

# ===========================================
# Stage 2: Build Go backend
# ===========================================
FROM golang:1.22-alpine AS backend-builder

WORKDIR /app
COPY backend/ ./backend/

RUN cd backend && CGO_ENABLED=0 go build -o /server .

COPY --from=frontend-builder /app/frontend/dist ./static

# ===========================================
# Stage 3: Final minimal image
# ===========================================
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=backend-builder /server .
COPY --from=backend-builder /static ./static

EXPOSE 10000

ENV PORT=10000

CMD ["./server"]
