# Dev-oriented Postgres with persistent data via bind mount (see database/.dockerignore).
# Example:
#   docker build -t recipes-db ./database
#   mkdir -p database/data
#   docker run --rm -p 5432:5432 -v "$(pwd)/database/data:/var/lib/postgresql/data" recipes-db

FROM pgvector/pgvector:pg16

ENV POSTGRES_DB=recipes \
    POSTGRES_HOST_AUTH_METHOD=trust

# Optional overrides at runtime: POSTGRES_USER, POSTGRES_PASSWORD, POSTGRES_DB

COPY db.sql /schema/db.sql
COPY docker-entrypoint.sh /usr/local/bin/recipes-db-entrypoint.sh
RUN chmod +x /usr/local/bin/recipes-db-entrypoint.sh

ENTRYPOINT ["/usr/local/bin/recipes-db-entrypoint.sh"]
CMD ["postgres"]
