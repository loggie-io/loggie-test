# Build the binary
FROM golang:1.17 as builder

# Copy in the go src
WORKDIR /
COPY . .
# Build
RUN go build -mod=vendor -o genfiles cmd/logfiles/genfiles.go

# Run
FROM debian:buster-slim
WORKDIR /
COPY --from=builder genfiles .

ENTRYPOINT ["/genfiles"]