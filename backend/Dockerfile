# Use the official Go image as the base image
FROM golang:1.20-alpine

# Set the working directory inside the container
WORKDIR /app

# Copy the Go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod tidy

# Copy the entire project into the container
COPY . .

# Build the Go application
RUN go build -o main cmd/server/main.go

# Expose the port the app will run on
EXPOSE 8080

# Command to run the Go application
CMD ["./main"]
