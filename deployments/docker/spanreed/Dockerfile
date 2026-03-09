FROM golang:1.22 AS build
WORKDIR /src
COPY . .
RUN go build -o /out/spanreed ./cmd/spanreed

FROM gcr.io/distroless/base-debian12
COPY --from=build /out/spanreed /app/spanreed
ENTRYPOINT ["/app/spanreed"]
