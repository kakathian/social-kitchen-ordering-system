FROM golang:1.15-alpine AS build_base

RUN apk add --no-cache git

# Set the Current Working Directory inside the container
WORKDIR /tmp/sharedkitchenordersystem-app

# We want to populate the module cache based on the go.{mod,sum} files.
COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

# Unit tests
# RUN CGO_ENABLED=0 go test -v

# Build the Go app
RUN go build -o ./out/sharedkitchenordersystem-app cmd/sharedkitchenordersystem/main.go

# Start fresh from a smaller image
FROM alpine:3.9 
RUN apk add ca-certificates

COPY --from=build_base /tmp/sharedkitchenordersystem-app/out/sharedkitchenordersystem-app /app/sharedkitchenordersystem-app

# This container exposes port 8080 to the outside world
EXPOSE 1323

# Run the binary program produced by `go install`
CMD ["/app/sharedkitchenordersystem-app"]