FROM golang:1.16-alpine3.13
WORKDIR /app
RUN apk add \
    --update \
    --no-cache \
    --repository http://dl-3.alpinelinux.org/alpine/edge/testing \
    --repository http://dl-3.alpinelinux.org/alpine/edge/main \
    vips-dev
COPY . .
RUN make build

FROM golang:1.16-alpine3.13
WORKDIR /app
COPY --from=0 /app/bin/imager /app/imager
CMD ["/app/imager"]  
