FROM golang:1.22 AS builder
WORKDIR /src
COPY go.mod go.work ./
COPY cmd/ cmd/
COPY internal/ internal/
RUN go build -o /out/stormlight ./cmd/stormlight

FROM gcr.io/distroless/base-debian12
COPY --from=builder /out/stormlight /usr/local/bin/stormlight
ENTRYPOINT ["/usr/local/bin/stormlight"]
