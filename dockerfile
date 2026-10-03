FROM golang:1.27-alpine AS build

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -o /servidor .

FROM scratch
COPY --from=build /servidor /servidor
USER 65532:65532
ENV SERVER_ADDR=:8080 TZ=America/Manaus
EXPOSE 8080
ENTRYPOINT ["/servidor"]
