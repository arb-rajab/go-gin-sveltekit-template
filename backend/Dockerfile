# Keep the Go version here equal to the `go` line in go.mod and to the
# go-version in .github/workflows/ci.yml (the CI pin may carry a newer patch).
FROM golang:1.26-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o starterapp . \
    && CGO_ENABLED=0 go build -o provision-operator ./cmd/provision-operator

FROM alpine:3.24

RUN adduser -D -u 10001 appuser
COPY --from=build /app/starterapp /usr/local/bin/starterapp
COPY --from=build /app/provision-operator /usr/local/bin/provision-operator

USER appuser
EXPOSE 8080

CMD ["starterapp"]
