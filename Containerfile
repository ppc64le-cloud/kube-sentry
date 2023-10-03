FROM ppc64le/golang:1.20-alpine as builder
RUN apk add sqlite-dev \
            sqlite-static \
            gcc \
            libc-dev
WORKDIR /kubertas
COPY . .
RUN CGO_ENABLED=1 go build -ldflags "-s -w -linkmode 'external' -extldflags '-static'" -v -o /kubertas/kubertas /kubertas/cmd

FROM scratch
COPY --from=builder /kubertas/kubertas /usr/bin/kubertas
WORKDIR /
CMD ["/usr/bin/kubertas"]
