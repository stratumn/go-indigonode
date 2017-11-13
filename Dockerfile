FROM stratumn/gobase:0.2.0

RUN addgroup -S -g 999 alice
RUN adduser -H -D -u 999 -G alice alice

RUN mkdir -p /usr/local/bin
COPY dist/linux-amd64/alice /usr/local/bin/

COPY docker/start-alice.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/start-alice.sh

RUN mkdir -p /usr/local/etc/alice
RUN chown alice:alice /usr/local/etc/alice/
RUN chmod 0700 /usr/local/etc/alice

RUN mkdir -p /usr/local/var/alice
RUN chown alice:alice /usr/local/var/alice/
RUN chmod 0700 /usr/local/var/alice

USER alice

CMD ["start-alice.sh"]

EXPOSE 5990 5991

VOLUME [ "/usr/local/etc/alice", "/usr/local/var/alice" ]
