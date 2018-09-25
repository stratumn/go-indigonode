FROM stratumn/gobase:latest

RUN addgroup -S -g 999 stratumn-node
RUN adduser -H -D -u 999 -G stratumn-node stratumn-node

RUN mkdir -p /usr/local/bin
COPY dist/linux-amd64/stratumn-node /usr/local/bin/

RUN mkdir -p /usr/local/etc/stratumn-node
RUN chown stratumn-node:stratumn-node /usr/local/etc/stratumn-node/
RUN chmod 0700 /usr/local/etc/stratumn-node

RUN mkdir -p /usr/local/var/stratumn-node
RUN chown stratumn-node:stratumn-node /usr/local/var/stratumn-node/
RUN chmod 0700 /usr/local/var/stratumn-node

USER stratumn-node

WORKDIR /usr/local/var/stratumn-node

EXPOSE 8903 8904 8905 8906
VOLUME [ "/usr/local/etc/stratumn-node", "/usr/local/var/stratumn-node" ]

ENTRYPOINT [ "/usr/local/bin/stratumn-node", "--core-config", "/usr/local/etc/stratumn-node/stratumn_node.core.toml", "--cli-config", "/usr/local/etc/stratumn-node/stratumn_node.cli.toml" ]