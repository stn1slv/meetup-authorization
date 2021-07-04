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
