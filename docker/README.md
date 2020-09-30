# building the docker image

```
make
make network
make run
```

Add to your docker-compose file:
```
networks:
  default:
    external:
      name: poolside-network
```