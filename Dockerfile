ARG GO_VERSION=1.24.5

# First stage: build the executable.
FROM golang:${GO_VERSION}-alpine AS build
RUN apk add --no-cache tzdata
ENV CGO_ENABLED=0
WORKDIR /src
COPY . .
RUN go mod download
RUN go build -ldflags "-s -w" -o web ./cmd/web


FROM scratch

COPY --from=build /usr/share/zoneinfo /usr/share/zoneinfo
ENV TZ=America/Chicago

WORKDIR /app

COPY --from=build /src/web .
COPY --from=build /src/tls/cert.pem /app/tls/cert.pem
COPY --from=build /src/tls/key.pem /app/tls/key.pem
COPY --from=build /src/.env .
COPY --from=build /src/ui/static/SharePics .

EXPOSE 443
ENTRYPOINT ["/app/web"]
