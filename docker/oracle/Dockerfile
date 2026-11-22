FROM golang:1.26.0-bookworm AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY oracle ./oracle
COPY smart-contract ./smart-contract
RUN CGO_ENABLED=0 go build -trimpath -o /bin/attestation-oracle ./oracle/cmd/oracle \
    && CGO_ENABLED=0 go build -trimpath -o /bin/deploy-private ./smart-contract/cmd/deploy-private \
    && CGO_ENABLED=0 go build -trimpath -o /bin/verify-private ./smart-contract/cmd/verify-private

FROM debian:bookworm-slim@sha256:734a72b3d6142a71993b4bc580ca01d8ca687f77a2c422d96a961d3376b179bc
RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates wget \
    && rm -rf /var/lib/apt/lists/*
COPY --from=build /bin/attestation-oracle /bin/deploy-private /bin/verify-private /usr/local/bin/
COPY --chmod=755 docker/oracle/development.sh /usr/local/bin/development
WORKDIR /app
ENTRYPOINT ["/usr/local/bin/development"]
