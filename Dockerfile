FROM scratch
MAINTAINER Qbox Inc.
ADD ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
ADD supergiant-api supergiant-api
EXPOSE 8080
ENTRYPOINT ["/supergiant-api"]
