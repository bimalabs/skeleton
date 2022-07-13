FROM ad3n/bima-cli:latest as builder

RUN apk update && apk add --no-cache git
RUN mkdir -p /go/src/app
WORKDIR /go/src/app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN bima dump && bima build bima

FROM alpine:latest

COPY --from=builder /go/src/app/bima /usr/local/bin/bima
COPY --from=builder /go/src/app/configs /usr/configs
COPY --from=builder /go/src/app/swaggers /usr/swaggers
RUN chmod a+x /usr/local/bin/bima
WORKDIR /usr

EXPOSE 7777

CMD ["/usr/local/bin/bima"]
