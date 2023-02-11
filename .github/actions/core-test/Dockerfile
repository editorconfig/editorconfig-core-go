FROM golang:bullseye

RUN set -xe && \
    apt-get update -y && \
    apt-get install -qy --no-install-recommends \
        cmake \
        git \
        make

CMD [ "make", "submodule", "test-core", "test-skipped" ]
