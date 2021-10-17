FROM alpine:3.14
COPY apcmetrics /usr/local/bin/apcmetrics
USER nobody
ENTRYPOINT ["apcmetrics"]
