FROM golang:1.23.5-alpine AS build

WORKDIR /app

# Устанавливаем необходимые пакеты для CGO
RUN apk add --no-cache build-base

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -o /forum-app ./cmd/web

FROM alpine:latest
WORKDIR /app

COPY --from=build /forum-app /app/
COPY ui /app/ui
COPY data /app/data

EXPOSE 4000
CMD ["./forum-app"]
