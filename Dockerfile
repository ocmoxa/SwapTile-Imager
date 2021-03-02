FROM golang:1.16-alpine3.13
WORKDIR /app
RUN apk add \
    --update \
    --no-cache \
    vips-dev \
    alpine-sdk
COPY . .
RUN make build

FROM golang:1.16-alpine3.13
WORKDIR /app
RUN apk add --update --no-cache vips
COPY --from=0 /app/bin/imager /app/imager
CMD ["/app/imager"]  
