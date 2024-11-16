FROM golang as builder

WORKDIR /app
COPY ./core /app/core
COPY ./examples /app/examples
COPY ./pkg /app/pkg
COPY ./go.mod /app/go.mod
COPY ./go.sum /app/go.sum

RUN go env -w GO111MODULE=on && go mod download

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/example /app/examples/main.go

FROM ubuntu:22.04 as base

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

WORKDIR /home/ubuntu

RUN apt-get update --yes && \
    apt-get upgrade --yes && \
    # Basic Utilities
    apt install --yes --no-install-recommends \
    bash \
    ca-certificates \
    curl \
    file \
    git \
    inotify-tools \
    jq \
    libgl1 \
    lsof \
    vim \
    tmux \
    procps \
    rsync \
    sudo \
    software-properties-common \
    unzip \
    wget \
    zip \
    graphviz && \
    # Build Tools and Development
    apt install --yes --no-install-recommends \
    build-essential \
    make \
    cmake \
    gfortran \
    libblas-dev \
    liblapack-dev && \
    apt-get autoremove -y && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/* && \
    # Set locale
    echo "en_US.UTF-8 UTF-8" > /etc/locale.gen

RUN  wget https://go.dev/dl/go1.23.3.linux-amd64.tar.gz
RUN  rm -rf /usr/local/go && tar -C /usr/local -xzf go1.23.3.linux-amd64.tar.gz
ENV  PATH="$PATH:/usr/local/go/bin"

COPY --from=builder /app/example /home/ubuntu/example

EXPOSE 38080
EXPOSE 38081
EXPOSE 38082

CMD ["/home/ubuntu/example"]
