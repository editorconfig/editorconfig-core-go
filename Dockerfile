FROM linuxkit/ca-certificates:v0.7

COPY editorconfig /usr/local/bin/

ENTRYPOINT [ "editorconfig" ]
