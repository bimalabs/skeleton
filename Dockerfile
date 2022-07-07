FROM ad3n/bima-cli:latest as builder

RUN mkdir -p /go/src/app
COPY . /go/src/app
WORKDIR /go/src/app
RUN cd /go/src/app && bima build bima

FROM alpine:latest

COPY --from=builder /go/src/app/bima /usr/local/bin/bima
RUN chmod a+x /usr/local/bin/bima

EXPOSE 7777

CMD ["/usr/local/bin/bima"]
