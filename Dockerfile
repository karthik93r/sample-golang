# Start with ubuntu 
FROM ubuntu:16.04

ARG DEBIAN_FRONTEND=noninteractive
RUN apt-get update && apt-get install --assume-yes apt-utils

# Install Git
RUN apt-get update && \
    apt-get install -y git-core && \
    git config --global user.name "Openn Build" && \
    git config --global user.email dev@openn.com.au 

# START install nvm and nodejs
##############################

# Replace shell with bash so we can source files
RUN rm /bin/sh && ln -s /bin/bash /bin/sh

# make sure apt is up to date
RUN apt-get update --fix-missing
RUN apt-get install -y curl
RUN apt-get install -y build-essential libssl-dev

ENV NVM_DIR /usr/local/nvm
ENV NODE_VERSION 8.8.1
ENV YARN_VERSION 1.3.2

# Install nvm , node, npm and yarn
RUN curl https://raw.githubusercontent.com/creationix/nvm/v0.33.6/install.sh | bash \
    && source $NVM_DIR/nvm.sh \
    && nvm install $NODE_VERSION \
    && nvm alias default $NODE_VERSION \
    && nvm use default \
    && curl -o- -L https://yarnpkg.com/install.sh | bash -s -- --version $YARN_VERSION

ENV NODE_PATH $NVM_DIR/v$NODE_VERSION/lib/node_modules
ENV PATH      $NVM_DIR/versions/node/v$NODE_VERSION/bin:/root/.yarn/bin:$PATH
# END install nvm and nodejs
##############################

RUN apt-get install -y wget

# START install golang
##############################
ENV GO_VERSION 1.11.3
RUN wget https://storage.googleapis.com/golang/go$GO_VERSION.linux-amd64.tar.gz -P /usr/local && \
    tar -C /usr/local -xzf /usr/local/go$GO_VERSION.linux-amd64.tar.gz

ENV PATH  $PATH:/usr/local/go/bin:/builds/go/bin
ENV GOPATH /builds/go
# END install golang
##############################

# Install Python (used by NewRelic)
RUN apt-get update && \
    apt-get install -y python3-pip && \
    apt-get install -y build-essential libssl-dev libffi-dev python-dev && \
    pip3 install --upgrade pip

# AWS cli installation
RUN apt-get update && \
    apt-get install -y awscli && \
    apt-get install -y zip && \
    pip3 install --upgrade awscli

#Install vi and atop
RUN apt-get install -y vim && \
    apt-get install -y atop

#Install bc, for time formatting
RUN apt-get update && \
    apt-get install bc

RUN mkdir -p /builds

COPY config.json /builds/config.json
COPY powerbi.go /builds/powerbi.go
COPY openn_build /openn_build

WORKDIR /