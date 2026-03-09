FROM golang:1.22 AS build
WORKDIR /src
COPY . .
RUN go build -o /out/radiant ./cmd/radiant

FROM gcr.io/distroless/base-debian12
COPY --from=build /out/radiant /app/radiant
ENTRYPOINT ["/app/radiant"]
