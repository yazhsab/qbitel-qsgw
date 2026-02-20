FROM rust:1.82-slim AS builder

RUN apt-get update && apt-get install -y pkg-config libssl-dev && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY Cargo.toml Cargo.lock ./
COPY types/ types/
COPY crypto/ crypto/
COPY tls/ tls/
COPY gateway/ gateway/

RUN cargo build --release --package quantun-qsgw-gateway

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*
COPY --from=builder /app/target/release/quantun-qsgw-gateway /usr/local/bin/qsgw-gateway

EXPOSE 8443
ENTRYPOINT ["qsgw-gateway"]
