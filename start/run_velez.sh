export VELEZ_PORT_GRPC=8080

docker network create verv

docker run \
  -d \
  --rm \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v /tmp/velez:/tmp/velez \
  -p ${VELEZ_PORT_GRPC}:80 \
  --name velez \
  godverv/velez:v0.1.34