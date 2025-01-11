# github.com/godverv/Velez

Velez is a lightweight working node manager.
It simply serves to starts/stops services on machine using docker.

If you are using verv ecosystem (rscli to build Go applications
and matreshka config to configurate applications) it opens more
robust experience.

Even if you're using only Velez to run containers it is still will
be much better experience than a manual setup (with docker-compose or presumably even a k8s)

# Not tested in production

For now. But with time this part will definitely change

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