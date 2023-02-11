FROM golang:bullseye

SHELL [ "/bin/bash", "-x", "-e", "-c" ]

ARG DEBIAN_FRONTEND=noninteraactive
RUN apt-get update -y && \
    apt-get install -qy --no-install-recommends \
        cmake \
        git \
        make

CMD [ "make", "test-core", "test-skipped" ]
