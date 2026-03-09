FROM golang:1.22 AS build
WORKDIR /src
COPY . .
RUN go build -o /out/stormlight ./cmd/stormlight

FROM gcr.io/distroless/base-debian12
COPY --from=build /out/stormlight /app/stormlight
ENTRYPOINT ["/app/stormlight"]
