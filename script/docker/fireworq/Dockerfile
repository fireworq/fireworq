# Dockerfile for fireworq
#
# $ docker build -t fireworq .
# $ docker run --rm fireworq

FROM golang:1.15

ARG FIREWORQ_ROOT
ENV FIREWORQ_ROOT=${FIREWORQ_ROOT}

RUN wget -O /usr/local/bin/wait-for-it.sh https://raw.githubusercontent.com/vishnubob/wait-for-it/master/wait-for-it.sh && \
    chmod +x /usr/local/bin/wait-for-it.sh && \
    go get github.com/tianon/gosu

WORKDIR "${FIREWORQ_ROOT}"

ENTRYPOINT [ "sh", "-c", "wait-for-it.sh -t 60 ${MYSQL_HOST}:${MYSQL_PORT} -- ${FIREWORQ_ROOT}/script/docker/fireworq/entrypoint.sh" ]
