 dockerfile
    FROM node:22-alpine AS frontend-builder

    WORKDIR /app/frontend

    COPY frontend/package.json frontend/package-lock.json ./
    RUN npm ci --ignore-scripts

    COPY frontend/ .
    RUN npm run build

    FROM golang:1.22-alpine AS backend-builder

    WORKDIR /app
    COPY backend/ ./backend/

    RUN cd backend && CGO_ENABLED=0 go build -o /server .

    COPY --from=frontend-builder /app/frontend/dist ./static

    FROM alpine:3.19

    RUN apk add --no-cache ca-certificates tzdata

    WORKDIR /app

    COPY --from=backend-builder /server .
    COPY --from=backend-builder /static ./static

    EXPOSE 10000
    ENV PORT=10000
    CMD ["./server"]
