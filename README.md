# go.vervstack.ru/Velez

Velez is a lightweight working node manager.
It simply serves to starts/stops services on machine using docker.

If you are using verv ecosystem (rscli to build Go applications
and matreshka config to configurate applications) it opens more
robust experience.

Even if you're using only Velez to run containers it is still will
be much better experience than a manual setup (with docker-compose or presumably even a k8s)

# Not tested in production

For now. But with time this part will definitely change

# Download and start

## Setup scripts

### Shell script

```shell
  wget -qO- scripts.verv.tech/init_node.sh | bash
```

### Docker run script

```shell
docker run \
   -p 53890:53890 \
   -d \
   --name velez \
   -v /var/run/docker.sock:/var/run/docker.sock \
   -v /dev/disk:/dev/disk \
   -v /run:/run \
   -v ~/velez:/tmp/velez \
   godverv/velez:latest
}
```

## Parameters

### Default port

You can access running Velez API with `<your-ip>:53890/api` by default

### Access key

Simply a sequence of random ASCII characters.
Must be used to connect to velez and execute API procedures

Key is stored at /tmp/velez/private.key

```shell
curl -X GET --location "http://<your-ip>:53890/api/version" \
    -H "Content-Type: application/json" \
-H "Grpc-Metadata-Velez-Auth: <your_private_key>"
```

It is better to store it in host system (or at least volume)
otherwise every restart private key will be refreshed

### Required volumes

```
-v /var/run/docker.sock:/var/run/docker.sock
-v /dev/disk:/dev/disk
-v /run:/run
```

These volumes are required for proper work of Velez.
You must provide them
# Features checklist

## Basic v1

- [x] Container manager Api
- [ ] Velez admin panel UI
- [ ] Vacuum
- [ ] Configuration synchronization

## Advanced

- [ ] Support Certificates and Basic Auth between nodes and services
- [ ] Human-readable errors parsing
- [ ] Secrets
- [ ] Automatic on-air configuration update

##### generated with love for coding by [RedSock CLI](https://github.com/Red-Sock/rscli)