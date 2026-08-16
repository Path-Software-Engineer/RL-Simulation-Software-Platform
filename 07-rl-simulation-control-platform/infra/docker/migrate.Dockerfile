FROM postgres:17-alpine

COPY database/migrations/ /migrations/
COPY infra/azure/run-migrations.sh /usr/local/bin/run-migrations

RUN chmod 0555 /usr/local/bin/run-migrations \
    && find /migrations -type f -name '*.sql' -exec chmod 0444 {} \;

USER postgres
ENTRYPOINT ["/usr/local/bin/run-migrations"]
