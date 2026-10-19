# The product service uses Elasticsearch; this legacy PostgreSQL image is not
# referenced by docker-compose.yaml.
FROM docker.elastic.co/elasticsearch/elasticsearch:8.15.0

ENV discovery.type=single-node
ENV xpack.security.enabled=false

CMD ["elasticsearch"]