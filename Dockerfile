FROM golang:1.16
WORKDIR /app
COPY . .
RUN make build

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=0 /app/bin/imager .
CMD ["/app/imager"]  
