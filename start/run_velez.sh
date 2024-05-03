export VELEZ_PORT_GRPC=50051
docker network rm verv
docker network create -d bridge verv

docker run \
  -d \
  --rm \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v /tmp/velez:/tmp/velez \
  -p ${VELEZ_PORT_GRPC}:80 \
  --name velez \
  --network verv \
  godverv/velez:v0.1.37