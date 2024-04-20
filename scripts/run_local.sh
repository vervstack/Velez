docker run \
   -v /var/run/docker.sock:/var/run/docker.sock \
   --name velez \
   -p 50052:53890 \
   --rm \
   -e VERV_NAME=velez \
   -e VELEZ_SHUT_DOWN_ON_EXIT=true \
  velez:local