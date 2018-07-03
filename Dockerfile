FROM golang:1.9-alpine

WORKDIR /go/src/github.com/justinbarrick/flux-operator

ADD vendor vendor
ADD version version
ADD cmd cmd
ADD pkg pkg

RUN go test github.com/justinbarrick/flux-operator/...
RUN CGO_ENABLED=0 go build -ldflags '-w -s' -a -installsuffix cgo -o flux-operator cmd/flux-operator/main.go

FROM scratch
COPY --from=0 /go/src/github.com/justinbarrick/flux-operator/flux-operator /flux-operator
ENTRYPOINT ["/flux-operator"]
