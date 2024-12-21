docker run \
   -v /var/run/docker.sock:/var/run/docker.sock \
   --name velez \
   -p 53891:53890 \
   -v /tmp/velez:/tmp/velez \
   -v verv:/opt/velez/smerds \
   -e VERV_NAME=velez \
   -e VELEZ_SHUT_DOWN_ON_EXIT=true \
   -e VELEZ_DISABLE_API_SECURITY=true \
  velez:local
