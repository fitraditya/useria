FROM golang:1.25-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /useria ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=build /useria ./useria
COPY --from=build /app/internal/database/migrations ./internal/database/migrations
COPY --from=build /app/templates ./templates
COPY --from=build /app/static ./static
# Run as two separate containers from this same image:
#   docker run ... useria tenant   (EXPOSE 8080)
#   docker run ... useria admin    (EXPOSE 9080)
EXPOSE 8080
EXPOSE 9080
ENTRYPOINT ["./useria"]
