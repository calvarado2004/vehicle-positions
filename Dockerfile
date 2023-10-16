# Atlanta's Vehicles Positions Dockerfile

# Start from the latest golang base image
FROM --platform=linux/amd64 docker.io/golang:latest as builder

# Add Maintainer Info
LABEL maintainer="Carlos Alvarado carlos-alvarado@outlook.com>"

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

COPY proto ./proto

COPY google_transit ./google_transit

COPY assets ./assets

COPY vehicles.html ./vehicles.html

COPY tripTypes.go ./tripTypes.go

COPY main.go ./main.go

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download


# Build the Go app
RUN CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -o vehicles-service .

RUN chmod +x /app/vehicles-service

FROM --platform=linux/amd64 docker.io/alpine:latest

COPY --from=builder /app /app

WORKDIR /app

CMD [ "./vehicles-service"]