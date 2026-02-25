FROM redis:7.4.0
COPY ./config/redis.conf.template /usr/local/etc/redis/redis.conf.template
COPY ./redis-entrypoint.sh /usr/local/bin/redis-entrypoint.sh
RUN chmod +x /usr/local/bin/redis-entrypoint.sh
RUN chmod -R 777 /data
RUN chmod -R 777 /var/log
RUN touch /var/log/redis.log
RUN chmod 777 /var/log/redis.log
EXPOSE 6379

ENTRYPOINT ["redis-entrypoint.sh"]
