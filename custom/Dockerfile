FROM debian:wheezy
RUN groupadd -r mygroup && useradd -r -g mygroup myuser
COPY ["./custom", "/custom"]
USER myuser
ENTRYPOINT ["/custom"]
