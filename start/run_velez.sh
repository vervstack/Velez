export VELEZ_PORT_GRPC=53890
#export IMAGE=godverv/velez:v0.1.43
export IMAGE=velez:local
docker network rm verv
docker network create -d bridge verv

docker run \
  -d \
  --rm \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v /tmp/velez:/tmp/velez \
  -p ${VELEZ_PORT_GRPC}:50051 \
  --name velez \
  --network verv \
  ${IMAGE}