FROM alpine:3.6

MAINTAINER Michael Dai "sarahdj0917@gmail.com"

ADD bin/luxun /
RUN mkdir /lib64 && \
    ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2 && \
    chmod a+x /luxun

CMD ["/luxun"]
