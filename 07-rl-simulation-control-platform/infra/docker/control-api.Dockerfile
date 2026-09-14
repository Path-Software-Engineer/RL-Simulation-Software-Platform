FROM golang:1.26.5-alpine AS build
WORKDIR /workspace
COPY services/control-api/go.mod services/control-api/go.sum ./services/control-api/
WORKDIR /workspace/services/control-api
RUN go mod download && go mod verify
COPY services/control-api/ .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/control-api ./cmd/api

FROM alpine:3.22.1
RUN apk add --no-cache ca-certificates wget && addgroup -S app && adduser -S app -G app
WORKDIR /app
COPY --from=build /out/control-api /app/control-api
COPY contracts/http/openapi.json /app/contracts/http/openapi.json
USER app
EXPOSE 8080
ENTRYPOINT ["/app/control-api"]
