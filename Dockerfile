FROM golang:1.15-alpine AS build_base

RUN apk add --no-cache git

# Set the Current Working Directory inside the container
WORKDIR /app/

# We want to populate the module cache based on the go.{mod,sum} files.
COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

ENV noOfOrdersToRead 10

# Unit tests
# RUN CGO_ENABLED=0 go test -v

# Build the Go app
RUN go build -o ./out/sharedkitchenordersystem-app cmd/sharedkitchenordersystem/main.go

# This container exposes port 8080 to the outside world
EXPOSE 1323

# Run the binary program produced by `go install`
# CMD ["/app/out/sharedkitchenordersystem-app"]
CMD ["sh", "-c","/app/out/sharedkitchenordersystem-app -noOfOrdersToRead=${noOfOrdersToRead}"]