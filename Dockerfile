FROM golang:1.10.1-alpine3.7

WORKDIR /go/src/github.com/justinbarrick/flux-operator

ADD vendor ./
ADD cmd pkg version ./

RUN go test github.com/justinbarrick/flux-operator/...
RUN CGO_ENABLED=0 go build -ldflags '-w -s' -a -installsuffix cgo -o flux-operator cmd/flux-operator/main.go

FROM scratch
COPY --from=0 /go/src/github.com/justinbarrick/flux-operator/flux-operator /flux-operator
ENTRYPOINT ["/flux-operator"]
