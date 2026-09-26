FROM golang:1.26.5-bookworm
WORKDIR /workspace
COPY go.mod go.sum ./
RUN go mod download
EXPOSE 8080 7331
