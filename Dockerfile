# Build Image
FROM golang:1.12 AS builder
# Consume required ENVs
ARG sshkey
ARG CI_PROJECT_NAME
# Set workdir
WORKDIR /$CI_PROJECT_NAME
# Copy all files
COPY . /$CI_PROJECT_NAME
# Add ssh key for private libs
RUN mkdir ~/.ssh/ && \
    chmod 700 ~/.ssh && \
    ssh-keyscan git.sstv.io > ~/.ssh/known_hosts && \
    echo "$sshkey" > ~/.ssh/id_rsa && \
    chmod 600 ~/.ssh/id_rsa
RUN git config --global url."git@git.sstv.io:".insteadOf "https://git.sstv.io/"
# Install dependencies
RUN go mod download
# Build binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o $CI_PROJECT_NAME ./cmd/serv/

# Image for certs
FROM alpine:latest as certs
# Get certs
RUN apk --update add ca-certificates

# Base Image
FROM alpine:latest
# Consume required ENVs
ARG CI_PROJECT_NAME
# Setup
WORKDIR /$CI_PROJECT_NAME
RUN apk --update add wkhtmltopdf
COPY --from=builder /$CI_PROJECT_NAME/$CI_PROJECT_NAME ./app
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
# Add dummy .env file
COPY .env.sample .env
COPY file .
# Run
CMD ./app
