FROM azazeal88/kentnetworkbuild as builder
RUN go get -d -v github.com/KentNetwork/kent-network-api
WORKDIR /go/src/github.com/KentNetwork/kent-network-api/.
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /go/src/github.com/KentNetwork/kent-network-api/app .
CMD ["./app"]
