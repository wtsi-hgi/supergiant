from alpine
RUN apk add --no-cache ca-certificates
ADD ./supergiant-api-linux-arm64 /supergiant-api
ADD ./config.json /
CMD ["/supergiant-api"]
