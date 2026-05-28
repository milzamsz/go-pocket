FROM golang:1.26.3-alpine3.23 AS builder

WORKDIR /src

RUN apk add --no-cache ca-certificates tzdata

COPY go.mod go.sum* ./
RUN go mod download

COPY . .

# Build prerequisites for clean-clone compiles (generated files are gitignored).
RUN go install github.com/a-h/templ/cmd/templ@v0.3.1020 \
  && templ generate \
  && go run ./cmd/cssbuild

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -trimpath -ldflags="-s -w" -o /out/go-pocket .

FROM alpine:3.23.3 AS runtime

WORKDIR /app

RUN addgroup -S app && adduser -S app -G app \
  && apk add --no-cache ca-certificates tzdata

COPY --from=builder /out/go-pocket /app/go-pocket
COPY --from=builder /src/.env.example /app/.env.example

USER app

EXPOSE 8090
VOLUME ["/app/pb_data"]

ENTRYPOINT ["/app/go-pocket"]
CMD ["serve", "--http=0.0.0.0:8090", "--dir=/app/pb_data"]
