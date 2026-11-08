FROM golang:1.27-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download 

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /app .

FROM alpine:3.22

RUN apk add --no-cache ca-certificates tzdata \
&& adduser -D -u 10001 app

USER app
COPY --from=build /app /app
ENV PORT=8080
EXPOSE 8080

CMD ["/app"]