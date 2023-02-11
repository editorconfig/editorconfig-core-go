FROM golang:bullseye

RUN set -xe \
 && apt-get update -y \
 &&  apt-get install -qy --no-install-recommends \
        cmake \
        git \
        make

VOLUME /github/workspace

CMD [ "make", "test-core", "test-skipped" ]
