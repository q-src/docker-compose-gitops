FROM golang:1.13 AS builder
WORKDIR /builder
COPY . .
ENV CGO_ENABLED=0
RUN go get -d -v . && \
    go build -o /gitops

FROM ixdotai/docker-compose:1.25.5
COPY --from=builder gitops /bin
VOLUME /git
VOLUME /ssh
ENV SSH_PRIV_KEY_FILE=/ssh/id_rsa
ENV SSH_KNOWN_HOSTS=/ssh/known_hosts
ENV REPOSITORY_PATH=/git
ENTRYPOINT ["gitops"]