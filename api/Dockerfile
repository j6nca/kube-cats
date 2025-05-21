FROM golang:1.21 as builder
WORKDIR /app
COPY . .
RUN go build -o kube-cat-api .

FROM gcr.io/distroless/static
COPY --from=builder /app/kube-cat-api /
ENTRYPOINT ["/kube-cat-api"]