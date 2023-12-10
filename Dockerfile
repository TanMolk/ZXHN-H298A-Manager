# Stage 1: build
FROM golang:1.21 AS builder
WORKDIR /app
COPY go.mod ./
COPY go.sum ./

RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o myapp .

# Stage 2: create runtime image
FROM alpine:latest

WORKDIR /root/

# copy result
COPY --from=builder /app/myapp .

CMD ["./myapp"]
