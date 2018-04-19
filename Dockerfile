FROM alpine:3.6

MAINTAINER Michael Dai "sarahdj0917@gmail.com"

ADD bin/luxun /
RUN chmod a+x /luxun

CMD ["/luxun"]
