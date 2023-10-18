FROM ppc64le/golang:1.20-alpine as builder
WORKDIR /kube-sentry
COPY . .
RUN go build -ldflags "-s -w" -v -o /kube-sentry/kube-sentry /kube-sentry/cmd

FROM scratch
COPY --from=builder /kube-sentry/kube-sentry /usr/bin/kube-sentry
WORKDIR /
CMD ["/usr/bin/kube-sentry"]
