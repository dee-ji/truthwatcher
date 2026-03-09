FROM golang:1.22 AS builder
WORKDIR /src
COPY go.mod go.work ./
COPY cmd/ cmd/
COPY internal/ internal/
RUN go build -o /out/radiant ./cmd/radiant

FROM gcr.io/distroless/base-debian12
COPY --from=builder /out/radiant /usr/local/bin/radiant
ENTRYPOINT ["/usr/local/bin/radiant"]
