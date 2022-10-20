# Demo case
The [presentation](presentation.pdf) from the meetup.

![Demo case](.img/demo-case.png?raw=true)
## Preparing
You have to add some entries to your ```/etc/hosts``` file:
```
127.0.0.1            keycloak
127.0.0.1            kafka
```
That's needed for host resolution because Kafka brokers and Kafka clients connecting to Keycloak have to use the same hostname to ensure the compatibility of generated access tokens. Also, when Kafka client connects to Kafka broker running inside docker image, the broker will redirect the client to ```kafka:9092```.

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
:warning: Before using Kafka console apps, you should add strimzi libs to your Kafka libs directory. To do this, clone [strimzi-kafka-oauth repository](https://github.com/strimzi/strimzi-kafka-oauth) and follow the steps on [the link](https://github.com/strimzi/strimzi-kafka-oauth#building). 
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
