# Dockerfile for fireworq
#
# $ docker build -t fireworq .
# $ docker run --rm fireworq

FROM golang:1.12

ARG FIREWORQ_ROOT
ENV FIREWORQ_ROOT=${FIREWORQ_ROOT}

ARG FIREWORQ_DEPS
ENV FIREWORQ_DEPS=${FIREWORQ_DEPS}

COPY Makefile glide.yaml glide.lock "${FIREWORQ_DEPS}/"

RUN wget -O /usr/local/bin/wait-for-it.sh https://raw.githubusercontent.com/vishnubob/wait-for-it/master/wait-for-it.sh && \
    chmod +x /usr/local/bin/wait-for-it.sh && \
    go get github.com/Masterminds/glide && \
    go get github.com/tianon/gosu && \
    mkdir -p "${FIREWORQ_DEPS}" && \
    cd "${FIREWORQ_DEPS}" && make clean && make deps

WORKDIR "${FIREWORQ_ROOT}"

ENTRYPOINT [ "sh", "-c", "wait-for-it.sh -t 60 ${MYSQL_HOST}:${MYSQL_PORT} -- ${FIREWORQ_ROOT}/script/docker/fireworq/entrypoint.sh" ]
