# Build Image
FROM golang:1.12 AS builder
# Consume required ENVs
ARG sshkey
ARG CI_PROJECT_NAME
# Set workdir
WORKDIR /$CI_PROJECT_NAME
# Copy go.mod and go.sum
COPY go.sum go.mod ./
# Add ssh key for private libs
RUN mkdir ~/.ssh/ && \
    chmod 700 ~/.ssh && \
    ssh-keyscan git.sstv.io > ~/.ssh/known_hosts && \
    echo "$sshkey" > ~/.ssh/id_rsa && \
    chmod 600 ~/.ssh/id_rsa
RUN git config --global url."git@git.sstv.io:".insteadOf "https://git.sstv.io/"
# Install dependencies
RUN go mod download
# Copy all files
COPY . .
# Build binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -mod=readonly -a -installsuffix cgo -o $CI_PROJECT_NAME ./cmd/serv/

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
RUN apk --update add wkhtmltopdf xvfb ttf-freefont fontconfig dbus
COPY --from=builder /$CI_PROJECT_NAME/$CI_PROJECT_NAME ./app
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY wkhtmltox.so.0.12.5 /usr/lib/libwkhtmltox.so.0.12.5
COPY wkhtmltopdf /usr/bin/wkhtmltopdf
COPY wkhtmltoimage /usr/bin/wkhtmltoimage
# Add dummy .env file
COPY .env.sample .env
COPY file file
# Run
CMD ./app
