FROM golang:bullseye

SHELL [ "/bin/bash", "-x", "-e" ]

RUN set -xe && \
    apt-get update -y && \
    apt-get install -qy --no-install-recommends \
        cmake \
        git \
        make

CMD [ "make", "test-core", "test-skipped" ]
