FROM alpine:3.6

RUN adduser -D flux-operator
USER flux-operator

ADD tmp/_output/bin/flux-operator /usr/local/bin/flux-operator
