## Start infrastructure
Cleanup:
```docker rm keycloak kafka zookeeper```

Go to docker directory of the repo:
```cd docker```

All the following docker-compose commands should be run from this directory.

Startup:
```docker-compose -f compose.yml -f keycloak/compose.yml -f keycloak-import/compose.yml -f kafka-oauth-strimzi/compose-authz.yml up --build```