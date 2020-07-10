FROM alpine:3.8

RUN apk add --no-cache ca-certificates

COPY dist/ecsy-*-linux /usr/local/bin/ecsy

ENTRYPOINT ["/usr/local/bin/ecsy"]
