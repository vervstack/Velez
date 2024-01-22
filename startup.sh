export VELEZ_PORT_GRPC=18506
export VELEZ_PORT_GW=18507
export VELEZ_KEY_PATH=/tmp/velez/private.key

docker run \
  -p ${VELEZ_PORT_GRPC}:50051 \
  -p ${VELEZ_PORT_GW}:50052 \
  -d \
  --name velez \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v /dev/disk:/dev/disk \
  -v /run:/run \
  -v ${VELEZ_KEY_PATH}:/tmp/velez/private.key \
  godverv/velez:latest