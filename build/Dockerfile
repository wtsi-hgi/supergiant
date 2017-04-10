# This Dockerfile creates the image for production
# (does not contain go and sources, only binaries)

FROM alpine:3.5

RUN apk add --no-cache \
        ca-certificates \
        tzdata

ENV KUBECTL_VERSION="1.5.6"
RUN apk add --no-cache curl \
 && kubectl_version_latest="$(curl --silent --url https://github.com/kubernetes/kubernetes/releases/latest | sed -rn 's/.*tag\/v([^\"]*).*/\1/p')" \
 && kubectl_version="${KUBECTL_VERSION:-$kubectl_version_latest}" \
 && echo "current kubectl version: $kubectl_version" \
 && echo "latest kubectl version: $kubectl_version_latest" \
 && curl \
        --url https://storage.googleapis.com/kubernetes-release/release/v${kubectl_version}/bin/linux/amd64/kubectl \
        --output /usr/bin/kubectl \
 && chmod +x /usr/bin/kubectl \
 && apk del curl \
 && ( [ $(stat -c %s /usr/bin/kubectl) -gt 500 ] || exit 1 )

RUN apk add --no-cache openssl \
 && wget https://pkg.cfssl.org/R1.2/cfssl_linux-amd64 -O /usr/bin/cfssl \
 && wget https://pkg.cfssl.org/R1.2/cfssljson_linux-amd64 -O /usr/bin/cfssljson \
 && chmod +x /usr/bin/cfssl* \
 && apk del openssl

EXPOSE 80 443
ENV SUPERGIANT_SQLITE_FILE="/data/.supergiant.db" \
    SUPERGIANT_PUBLISH_HOST="0.0.0.0" \
    SUPERGIANT_HTTP_PORT="80" \
    SUPERGIANT_HTTPS_PORT="443" \
    SUPERGIANT_SSL_CERT_FILE="/data/ssl/cert.pem" \
    SUPERGIANT_SSL_KEY_FILE="/data/ssl/key.pem" \
    SUPERGIANT_UI_ENABLED="true" \
    SUPERGIANT_LOG_FILE="/dev/stdout" \
    SUPERGIANT_LOG_LEVEL="debug" \
    SUPERGIANT_DATA_DIR="/data" \
    TIMEZONE="Etc/UTC"

COPY build/entrypoint.sh /
COPY build/bin/* /usr/local/bin/
COPY build/dist/bin /usr/bin
COPY build/dist/config-example /etc/supergiant-example

ENTRYPOINT ["/entrypoint.sh"]
CMD ["/usr/local/bin/supergiant-server-wrapper"]
