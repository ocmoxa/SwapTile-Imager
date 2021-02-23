FROM golang:1.16
WORKDIR /app
COPY . .
RUN make build

FROM golang:1.16
WORKDIR /app
COPY --from=0 /app/bin/imager /app/imager
CMD ["/app/imager"]  
