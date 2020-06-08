FROM golang:1.14.4 as builder
ENV APP_DIR /go/src/github.com/fireworq/fireworq

WORKDIR ${APP_DIR}
COPY . .
RUN make release

FROM alpine:3.12.0
ENV APP_DIR /go/src/github.com/fireworq/fireworq

COPY --from=builder ${APP_DIR}/fireworq /usr/local/bin/
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/fireworq"]
