FROM ubuntu:jammy
RUN apt update -qqy && apt install -qqy ca-certificates
COPY dist/gacr /usr/bin/gacr
RUN mkdir -p /var/lib/gacr
COPY example/example.conf /var/lib/gacr/
COPY example/scripts/* /var/lib/gacr/
CMD ["gacr", "--config", "/var/lib/gacr/example.conf", "--once"]
