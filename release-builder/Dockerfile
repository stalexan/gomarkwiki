FROM debian:latest

ARG GO_VERSION=1.20.4

# Install packages.
USER root
RUN set -eux; \
    apt-get update; \
    apt-get install -y --no-install-recommends \
        bzip2 \
        ca-certificates \
        git \
        gnupg \
        wget \
        zip; \
    rm -rf /var/lib/apt/lists/* ;

# Install Go.
RUN set -eux; \
    GO_URL="https://dl.google.com/go/go${GO_VERSION}.linux-amd64.tar.gz"; \
    wget -O go.tar.gz.asc "${GO_URL}.asc"; \
    wget -O go.tar.gz "$GO_URL" --progress=dot:giga; \
    GNUPGHOME="$(mktemp -d)"; export GNUPGHOME; \
    gpg --batch --keyserver keyserver.ubuntu.com --recv-keys 'EB4C 1BFD 4F04 2F6D DDCC  EC91 7721 F63B D38B 4796'; \
    gpg --batch --verify go.tar.gz.asc go.tar.gz; \
    gpgconf --kill all; \
    rm -rf "$GNUPGHOME" go.tgz.asc; \
    tar -C /usr/local -xzf go.tar.gz; \
    rm go.tar.gz;

ENV \
    # Add go and local binaries to PATH.
    PATH="/usr/local/go/bin:/usr/local/bin:${PATH}" \
    # Set default timezone.
    TZ=America/New_York \
    # Configure user home.
    HOME=/home/build \
    # Disable cgo.
    CGO_ENABLED=0

# Add build user.
RUN set -eux; \
    useradd -m -d $HOME -s /bin/bash build

# Create build directories.
RUN set -eux; \
    mkdir /gomarkwiki && chown build /gomarkwiki; \
    mkdir /output && chown build /output

# Run everything below as user build.
USER build
WORKDIR /gomarkwiki

CMD bash
