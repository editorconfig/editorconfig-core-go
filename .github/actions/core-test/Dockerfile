FROM golang:bullseye

RUN set -xe && \
    apt-get update -y && \
    apt-get install -qy --no-install-recommends \
        cmake \
        git \
        make \
 && git config --global --add safe.directory /github/workspace

VOLUME /github/workspace

CMD [ "make", "submodule", "test-core", "test-skipped" ]
