# Start infrastructure
### Cleanup

```
docker rm keycloak kafka zookeeper
```

### Startup
Go to docker directory of the repo:

```
cd docker
```

All the following docker-compose commands should be run from this directory.

Starting infra:

```
docker-compose -f compose.yml -f keycloak/compose.yml -f keycloak-import/compose.yml -f kafka-oauth-strimzi/compose-authz.yml up --build
```

### Testing 

#### Console tools
Produce:
```
kafka-console-producer.sh --broker-list kafka:9092 --topic a_messages --producer.config=service-a.properties
```
Consume:
```
kafka-console-consumer.sh --bootstrap-server kafka:9092 --topic a_messages --from-beginning --consumer.config=service-b.properties --group a_consumer_group_1
```

