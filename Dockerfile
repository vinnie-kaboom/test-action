FROM golang:1.21 AS builder

WORKDIR /app
COPY . .
RUN go build -o /app/action

FROM ubuntu:22.04

# Install Python, Ansible and required packages
RUN apt-get update && \
    apt-get install -y python3 python3-pip openssh-client && \
    apt-get clean && \
    pip3 install ansible

# Copy the compiled Go binary
COPY --from=builder /app/action /usr/local/bin/action

# Set the entrypoint
ENTRYPOINT ["/usr/local/bin/action"] 