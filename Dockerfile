FROM golang:1.15.0 as builder
ENV APP_DIR /go/src/github.com/fireworq/fireworq

WORKDIR ${APP_DIR}
COPY . .
RUN make release PRERELEASE=

FROM alpine:3.12.0
ENV APP_DIR /go/src/github.com/fireworq/fireworq

COPY --from=builder ${APP_DIR}/fireworq /usr/local/bin/
ENV FIREWORQ_BIND 0.0.0.0:8080
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/fireworq"]
