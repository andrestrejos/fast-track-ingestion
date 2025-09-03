FROM golang:1.25 AS builder
WORKDIR /src


COPY go.mod ./
RUN go mod download


COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/app ./cmd/app


FROM gcr.io/distroless/base-debian12
WORKDIR /
COPY --from=builder /out/app /app
ENTRYPOINT ["/app"]

