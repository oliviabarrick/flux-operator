FROM scratch
COPY ./flux-operator /flux-operator
ENTRYPOINT ["/flux-operator"]
