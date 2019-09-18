FROM redis:5.0.5

COPY docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh
COPY cluster-config.sh /usr/local/bin/cluster-config.sh
COPY redis.conf /data/

RUN chown -R redis:redis /data
RUN chmod +x /usr/local/bin/*.sh

ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]

EXPOSE 6379
CMD ["redis-server"]
