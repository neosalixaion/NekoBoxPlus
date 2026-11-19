# syntax=docker/dockerfile:1.7

ARG DEBIAN_VERSION=bookworm
ARG GO_VERSION=1.27.3

# Use the official Go image as the bootstrap compiler. GitHub Actions runners
# can pull it from Docker Hub even when direct Google Storage downloads are
# denied to containers; rebuild the patched runtime from the matching source
# tag so the project's runtime patches are still applied.
FROM golang:${GO_VERSION}-bookworm AS go-builder

ARG GO_VERSION=1.27.3
ARG GO_PATCH_DIR=buildScript/lib/core/go-runtime-patches

ENV GOROOT_BOOTSTRAP=/usr/local/go

RUN apt-get update && apt-get install -y --no-install-recommends \
      ca-certificates git patch bash make gcc g++ libc6-dev pkg-config \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /tmp/go-src

RUN git clone --depth 1 --branch "go${GO_VERSION}" https://github.com/golang/go.git .

COPY ${GO_PATCH_DIR} /tmp/go-runtime-patches

RUN set -eux; \
    for patch_file in /tmp/go-runtime-patches/*.patch; do \
      echo "Applying ${patch_file}"; \
      patch --batch --fuzz=0 -p1 < "${patch_file}"; \
    done; \
    cd src; \
    ./make.bash; \
    cd /; \
    rm -rf /root/.cache /tmp/go-runtime-patches

FROM debian:${DEBIAN_VERSION}

ARG GO_VERSION=1.27.3
ARG ADBLOCK_RUST_TOOLCHAIN=1.97.0

ENV GO_VERSION=${GO_VERSION} \
    GOROOT=/usr/local/go \
    GOPATH=/go \
    GOMODCACHE=/go/pkg/mod \
    GOCACHE=/go-cache \
    ANDROID_HOME=/opt/android-sdk \
    ANDROID_SDK_ROOT=/opt/android-sdk \
    JAVA_HOME=/usr/lib/jvm/java-17-openjdk-amd64 \
    CGO_ENABLED=1 \
    GO386=softfloat \
    PATH=/root/.cargo/bin:/usr/local/go/bin:/go/bin:/opt/android-sdk/cmdline-tools/latest/bin:/opt/android-sdk/platform-tools:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin

RUN apt-get update && apt-get install -y --no-install-recommends \
      bash ca-certificates curl file git make patch unzip zip \
      gcc g++ libc6-dev pkg-config openssh-client \
      nodejs npm \
      openjdk-17-jdk-headless \
    && rm -rf /var/lib/apt/lists/* \
    && mkdir -p /workspace /go /go-cache /opt/android-sdk \
    && chmod -R 0777 /go /go-cache /workspace

COPY --from=go-builder /tmp/go-src /usr/local/go

RUN set -eux; \
    curl -fsSL https://sh.rustup.rs | sh -s -- -y --profile minimal --default-toolchain "${ADBLOCK_RUST_TOOLCHAIN}"; \
    rustup target add \
      armv7-linux-androideabi \
      aarch64-linux-android \
      i686-linux-android \
      x86_64-linux-android; \
    cargo --version; \
    rustc --version; \
    node --version; \
    npm --version

RUN go version && which go && go env GOROOT GOPATH

# The wrapper always bind-mounts the installed host Android SDK at
# /opt/android-sdk, so downloading a second copy inside this image is wasteful.

WORKDIR /workspace/project

CMD ["bash", "buildScript/lib/core.sh"]
