FROM golang:1.22 AS build
WORKDIR /src
COPY . .
RUN go build -o /out/tw-server ./cmd/tw-server

FROM gcr.io/distroless/base-debian12
COPY --from=build /out/tw-server /app/tw-server
ENTRYPOINT ["/app/tw-server"]
