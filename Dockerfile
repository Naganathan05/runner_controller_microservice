# golang 1.23
FROM golang:1.23

# Set the Current Working Directory inside the container
WORKDIR /runner_controller_microservice

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN go build -o main .

# This container exposes port 5002 to the outside world
EXPOSE 5002

# Command to run the executable
CMD ["./main"]
