FROM alpine:3.20

COPY editorconfig /usr/local/bin/

ENTRYPOINT [ "editorconfig" ]
