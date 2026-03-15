FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /bin/server ./cmd

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
COPY --from=builder /bin/server /bin/server
EXPOSE 8083
CMD ["/bin/server"]
