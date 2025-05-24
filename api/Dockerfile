FROM golang:1.24 AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o kube-cat-api

FROM gcr.io/distroless/static:nonroot
COPY --from=builder /app/kube-cat-api /
ENTRYPOINT ["/kube-cat-api"]