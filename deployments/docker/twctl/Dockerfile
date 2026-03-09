FROM golang:1.22 AS builder
WORKDIR /src
COPY go.mod go.work ./
COPY cmd/ cmd/
COPY internal/ internal/
RUN go build -o /out/twctl ./cmd/twctl

FROM gcr.io/distroless/base-debian12
COPY --from=builder /out/twctl /usr/local/bin/twctl
ENTRYPOINT ["/usr/local/bin/twctl"]
