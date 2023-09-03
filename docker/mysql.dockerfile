FROM mysql:8.0

RUN mkdir -p /var/log/mysql
RUN chown mysql:mysql /var/log/mysql
