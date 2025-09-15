FROM alpine:3.8

# alpine image no /etc/nsswitch.conf, dns resolve order is not /etc/hosts first
RUN echo 'hosts: files mdns4_minimal [NOTFOUND=return] dns mdns4' > /etc/nsswitch.conf

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.tuna.tsinghua.edu.cn/g' /etc/apk/repositories
RUN apk add --update ca-certificates tzdata && rm -rf /var/cache/apk/* /tmp/*
RUN mkdir -p /etc/ssl/certs/ && update-ca-certificates --fresh

ADD bin/ /app/
ADD dashboard/ /app/dashboard

WORKDIR /app

CMD ["/app/lumina"]
