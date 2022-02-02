# syntax=docker/dockerfile:1
FROM alpine
VOLUME /export
ENTRYPOINT ["/iauditor-exporter"]
COPY iauditor-exporter /
