
docker run \
  --rm \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -p 8090:53890 \
  --name velez \
  --network verv godverv/velez:v0.1.15