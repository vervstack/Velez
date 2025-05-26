export VELEZ_PORT=53890
export VELEZ_KEY_PATH=~/velez
export MATRESHKA_PASS=

run_velez() {
  local velezPort=$1
  local velezKeyPath=$2

  docker run \
    -p "${velezPort}":53890 \
    -e MATRESHKA_PASS="${MATRESHKA_PASS}" \
    -d \
    --name velez \
    -v /var/run/docker.sock:/var/run/docker.sock \
    -v "${velezKeyPath}":/tmp/velez \
    vervstack/velez:v0.1.72
}

run_velez $VELEZ_PORT $VELEZ_KEY_PATH
