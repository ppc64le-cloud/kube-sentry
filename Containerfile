FROM ppc64le/golang:1.20-alpine as builder
RUN apk add sqlite-dev \
            sqlite-static \
            gcc \
            libc-dev
WORKDIR /kube-sentry
COPY . .
RUN CGO_ENABLED=1 go build -ldflags "-s -w -linkmode 'external' -extldflags '-static'" -v -o /kube-sentry/kube-sentry /kube-sentry/cmd

FROM scratch
COPY --from=builder /kube-sentry/kube-sentry /usr/bin/kube-sentry
WORKDIR /
CMD ["/usr/bin/kube-sentry"]
