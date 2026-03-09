FROM golang:1.22 AS build
WORKDIR /src
COPY . .
RUN go build -o /out/squire ./cmd/squire

FROM gcr.io/distroless/base-debian12
COPY --from=build /out/squire /app/squire
ENTRYPOINT ["/app/squire"]
