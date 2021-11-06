FROM golang:latest as api-builder
WORKDIR  /src/github.com/dinumathai/admission-webhook-sample/
COPY . /src/github.com/dinumathai/admission-webhook-sample/
RUN CGO_ENABLED=0 go install github.com/dinumathai/admission-webhook-sample

FROM alpine:latest

# COPY application to workdir
WORKDIR /
COPY --from=api-builder /go/bin/admission-webhook-sample admission-webhook-sample

RUN chmod a+x /admission-webhook-sample

# Now tell Docker what command to run when the container starts...
CMD ["/admission-webhook-sample", "-stderrthreshold=INFO","-v=3"]
