FROM golang:1.22 AS build
WORKDIR /src
COPY . .
RUN go build -o /out/tw-worker ./cmd/tw-worker

FROM gcr.io/distroless/base-debian12
COPY --from=build /out/tw-worker /app/tw-worker
ENTRYPOINT ["/app/tw-worker"]
