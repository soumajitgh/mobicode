FROM golang:1.26-bookworm AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 go build -o /out/api ./cmd/api

FROM debian:bookworm-slim
RUN useradd --system --create-home appuser
WORKDIR /app
COPY --from=build /out/api /app/api
RUN mkdir /data && chown appuser:appuser /data
USER appuser
ENV SHARED_ENV=production SERVER_LOG_DEVELOPMENT=false SERVER_LOG_COLOR=false
EXPOSE 8080
ENTRYPOINT ["/app/api"]
