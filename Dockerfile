FROM stratumn/gobase:latest

RUN addgroup -S -g 999 indigo-node
RUN adduser -H -D -u 999 -G indigo-node indigo-node

RUN mkdir -p /usr/local/bin
COPY dist/linux-amd64/indigo-node /usr/local/bin/

RUN mkdir -p /usr/local/etc/indigo-node
RUN chown indigo-node:indigo-node /usr/local/etc/indigo-node/
RUN chmod 0700 /usr/local/etc/indigo-node

RUN mkdir -p /usr/local/var/indigo-node
RUN chown indigo-node:indigo-node /usr/local/var/indigo-node/
RUN chmod 0700 /usr/local/var/indigo-node

USER indigo-node

WORKDIR /usr/local/var/indigo-node

EXPOSE 8903 8904 8905 8906
VOLUME [ "/usr/local/etc/indigo-node", "/usr/local/var/indigo-node" ]

ENTRYPOINT [ "/usr/local/bin/indigo-node", "--core-config", "/usr/local/etc/indigo-node/indigo_node.core.toml", "--cli-config", "/usr/local/etc/indigo-node/indigo_node.cli.toml" ]