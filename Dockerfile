# Use the official Alpine Linux as a parent image
FROM golang:alpine

# Set the working directory inside the container
WORKDIR /app

# Copy the local source files to the container's working directory
COPY . .

# Build the Go application inside the container
RUN go build -o main

# Run the Go application when the container starts
CMD ["./main"]