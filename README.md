# Demo case
![Demo case](.img/demo-case.png?raw=true)
## Start infrastructure
#### Cleanup

```
docker rm keycloak kafka zookeeper
```

#### Startup
Go to docker directory of the repo:

```
cd docker
```

All the following docker-compose commands should be run from this directory.

Starting infra:

```
docker-compose -f compose.yml -f keycloak/compose.yml -f keycloak-import/compose.yml -f kafka-oauth-strimzi/compose-authz.yml up --build
```

## Run and test 

#### Console tools
###### Produce
Service-A
```
kafka-console-producer.sh --broker-list kafka:9092 --topic a_messages --producer.config=service-a.properties
```
Service-C
```
kafka-console-producer.sh --broker-list kafka:9092 --topic a_messages --producer.config=service-c.properties
```

###### Consume
Service-B
```
kafka-console-consumer.sh --bootstrap-server kafka:9092 --topic a_messages --from-beginning --consumer.config=service-b.properties --group a_consumer_group_1
```
#### Run demo apps
###### Service-A
```
cd service-a
mvn spring-boot:run
```
###### Service-B
```
cd service-b
mvn spring-boot:run
```
###### Service-C
```
cd service-c
go run main.go
```
###### Service-D
```
cd service-d
go run main.go
```
