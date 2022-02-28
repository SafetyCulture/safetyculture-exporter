# syntax=docker/dockerfile:1
FROM alpine
VOLUME /export
VOLUME /config
ENTRYPOINT ["./iauditor-exporter"]
COPY iauditor-exporter /
