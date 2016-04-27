## Dockerfile for Earthquake
## Available at Docker Hub: osrg/earthquake
FROM osrg/dind-ovs-ryu
MAINTAINER Akihiro Suda <suda.akihiro@lab.ntt.co.jp>

RUN apt-get update && apt-get install -y --no-install-recommends \
    ## Install Earthquake deps
    protobuf-compiler pkg-config libzmq3-dev libnetfilter-queue-dev \
    ## (Optional) Install Java inspector deps
    default-jdk maven \
    ## (Optional) Install useful stuffs
    sudo ant ant-optional \
    ## (Optional) Install MongoDB storage
    mongodb \
    ## (Optional) Install FUSE inspector deps
    fuse \
    ## (Optional) Install pynmz deps
    python-flask python-scapy python-zmq \
    ## (Optional) Install pynmz nfqhook deps
    libnetfilter-queue1 python-prctl

## Install Go 1.6
RUN curl https://storage.googleapis.com/golang/go1.6.linux-amd64.tar.gz | tar Cxz /usr/local && mkdir /gopath
ENV PATH /usr/local/go/bin:$PATH
ENV GOPATH /gopath

## (Optional) Install pynmz deps
RUN pip install hexdump

## (Optional) Install hookswitch
RUN pip install hookswitch==0.0.2

## (Optional) Install pipework for DinD
RUN curl https://raw.githubusercontent.com/jpetazzo/pipework/master/pipework -o /usr/local/bin/pipework
RUN chmod +x /usr/local/bin/pipework

## (Optional) Create a user for nfqueue sandbox
RUN useradd -m nfqhooked

## Copy Earthquake to /earthquake
ADD . /earthquake
WORKDIR /earthquake
RUN ( git submodule init && git submodule update )
ENV PYTHONPATH /earthquake:$PYTHONPATH

## Build Earthquake
RUN ./build

## Silence dind logs
ENV LOG file

## Start init (does NOT enable DinD/OVS/Ryu by default)
ADD misc/docker/eq-init.py /eq-init.py
CMD ["/eq-init.py"]
