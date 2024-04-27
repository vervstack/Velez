export VELEZ_PORT_GRPC=53890

docker network create verv

docker run \
  -d \
  --rm \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v /tmp/velez:/tmp/velez \
  -p ${VELEZ_PORT_GRPC}:53890 \
  --name velez \
  --network verv \
  godverv/velez:v0.1.16