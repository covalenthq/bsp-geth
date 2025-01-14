# Support setting various labels on the final image
ARG COMMIT=""
ARG VERSION=""
ARG BUILDNUM=""

# Build Geth in a stock Go builder container
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache gcc musl-dev linux-headers git

# Get dependencies - will also be cached if we won't change go.mod/go.sum
COPY go.mod /go-ethereum/
COPY go.sum /go-ethereum/
RUN cd /go-ethereum && go mod download

ADD . /go-ethereum
RUN cd /go-ethereum && go run build/ci.go install -static ./cmd/geth

# Pull Geth into a second stage deploy alpine container
FROM alpine:latest

RUN apk add --no-cache ca-certificates
COPY --from=builder /go-ethereum/build/bin/geth /usr/local/bin/

EXPOSE 8545 8546 30303 30303/udp
ENTRYPOINT ["geth", "--mainnet", "--syncmode", "full", "--datadir", "/root/.ethereum/covalent", "--replication.targets", "redis://localhost:6379/?topic=replication", "--replica.result", "true", "--replica.specimen", "true", "--replica.blob", "true"]

# Add some metadata labels to help programmatic image consumption
ARG COMMIT=""
ARG VERSION=""
ARG BUILDNUM=""

LABEL commit="$COMMIT" version="$VERSION" buildnum="$BUILDNUM"
