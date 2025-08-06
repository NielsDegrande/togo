FROM golang:1.24.5-alpine

LABEL NAME=togo
LABEL VERSION=1.0.0

WORKDIR /app

# Install system dependencies.
RUN apk add --no-cache \
    bash \
    git \
    python3 \
    uv

# Create a virtual environment.
RUN uv venv /opt/venv
ENV VIRTUAL_ENV=/opt/venv

# Install pre-commit.
RUN uv pip install pre-commit

# Install Go tools required by pre-commit hooks.
RUN go install mvdan.cc/gofumpt@latest && \
    go install golang.org/x/tools/cmd/goimports@latest && \
    go install github.com/securego/gosec/v2/cmd/gosec@latest

# Install golangci-lint
RUN wget -O- -nv https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin

# Copy go.mod and go.sum first for better caching.
COPY go.mod go.sum* ./

# Download dependencies.
RUN go mod download

# Copy the pre-commit config.
COPY .pre-commit-config.yaml .pre-commit-config.yaml

# Set up git configuration for testuser.
RUN git config --global --add safe.directory /app && \
    git config --global --add safe.directory '*'

ENTRYPOINT ["sh"]
