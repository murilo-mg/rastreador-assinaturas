FROM golang:1.22-alpine AS build

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o servidor .

FROM alpine:latest

WORKDIR /app
COPY --from=build /app/servidor .

EXPOSE 8080
CMD ["./servidor"]