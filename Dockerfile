FROM alpine:3.21

COPY editorconfig /usr/local/bin/

ENTRYPOINT [ "editorconfig" ]
