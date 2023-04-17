FROM golang:1.20.1

## Install unzip and jq
RUN apt-get update && apt-get install -y --no-install-recommends \
    unzip \
    jq

# Install protobuf from releas page
RUN PB_REL="https://github.com/protocolbuffers/protobuf/releases" && VERSION=22.0 \
    && curl -LO $PB_REL/download/v$VERSION/protoc-$VERSION-linux-x86_64.zip \
    && unzip protoc-$VERSION-linux-x86_64.zip -d $HOME/.local \
    && export PATH="$PATH:$HOME/.local/bin" \
    && echo 'export PATH=$PATH:$HOME/.local/bin' >> ~/.bashrc

# Install LLVM as default compiler, for memory sanitizer

## Add apt key for LLVM repository
#RUN wget -O - https://apt.llvm.org/llvm-snapshot.gpg.key | apt-key add -
RUN curl -sS https://apt.llvm.org/llvm-snapshot.gpg.key | apt-key add -

## Add LLVM apt repository
RUN echo "deb http://apt.llvm.org/stretch/ llvm-toolchain-stretch-5.0 main" | tee -a /etc/apt/sources.list

## Install clang from LLVM repository
RUN apt-get update && apt-get install -y --no-install-recommends \
    apt-utils \
    apt-transport-https \
    clang \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

## Set Clang as default CC
ENV set_clang /etc/profile.d/set-clang-cc.sh
RUN echo "export CC=clang-5.0" | tee -a ${set_clang} && chmod a+x ${set_clang}

# Sonar Scanner
RUN mkdir -p /progs && \
    # SonarScanner
    cd /progs && \
    wget https://binaries.sonarsource.com/Distribution/sonar-scanner-cli/sonar-scanner-cli-4.3.0.2102-linux.zip && \
    unzip sonar-scanner-cli-4.3.0.2102-linux.zip && \
    mv sonar-scanner-4.3.0.2102-linux /opt/sonar-scanner && \
    echo 'export PATH=$PATH:/opt/sonar-scanner/bin' >> ~/.bashrc && \
    . ~/.bashrc

# JFrog Client
RUN curl -fL https://getcli.jfrog.io/v2 | sh && \
    mv jfrog /usr/bin/jfrog

# Install Deb/Rpm packaging tool 'goreleaser'
RUN echo 'deb [trusted=yes] https://repo.goreleaser.com/apt/ /' | tee /etc/apt/sources.list.d/goreleaser.list
RUN apt update
RUN apt install goreleaser -y

# Setup Go environment
ENV GOPATH /go
ENV PATH ${GOPATH}/bin:$PATH

# Download go modules
ARG BASE_PACKAGE_NAME
WORKDIR /go/src/$BASE_PACKAGE_NAME/
COPY go.mod .
COPY go.sum .
RUN GO111MODULE=on go mod download

# Install golint
RUN go get -d -v -u golang.org/x/lint/golint

# Install gRPC & Protobuf dependencies
RUN go get -d -v -u google.golang.org/grpc
RUN go get -d -v -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway
RUN go get -d -v -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger
RUN go get -d -v -u github.com/golang/protobuf/protoc-gen-go

COPY . /go/src/$BASE_PACKAGE_NAME/
