FROM golang:1.22-alpine AS build

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o notes-app

# final minimal image
FROM alpine:latest
RUN apk add --no-cache ca-certificates
WORKDIR /root/

# Copy everything from the build stage
COPY --from=build /app .

EXPOSE 60
CMD ["./notes-app"]