FROM alpine:3.11

RUN apk add --no-cache ca-certificates

COPY dist/ecsy-*-linux /usr/local/bin/ecsy

RUN chmod +x /usr/local/bin/ecsy

CMD ["/usr/local/bin/ecsy"]
