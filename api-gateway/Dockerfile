FROM golang:1.24-alpine AS builder 
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o gateway ./cmd/main.go


FROM golang:1.24-alpine 
WORKDIR /app
COPY --from=builder /app/gateway .
EXPOSE 8000
CMD ["./gateway"]