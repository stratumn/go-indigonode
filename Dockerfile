FROM stratumn/gobase:0.3.0

RUN addgroup -S -g 999 alice
RUN adduser -H -D -u 999 -G alice alice

RUN mkdir -p /usr/local/bin
COPY dist/linux-amd64/alice /usr/local/bin/

RUN mkdir -p /usr/local/etc/alice
RUN chown alice:alice /usr/local/etc/alice/
RUN chmod 0700 /usr/local/etc/alice

RUN mkdir -p /usr/local/var/alice
RUN chown alice:alice /usr/local/var/alice/
RUN chmod 0700 /usr/local/var/alice

USER alice

WORKDIR /usr/local/var/alice

ENTRYPOINT [ "/usr/local/bin/alice" ]

EXPOSE 8903 8904 8905 8906

VOLUME [ "/usr/local/etc/alice", "/usr/local/var/alice" ]
