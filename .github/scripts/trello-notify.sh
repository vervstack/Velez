#!/bin/sh
# Tells the Artel tracts which Trello cards a release touches. A card is named by a
# https://trello.com/c/<shortLink> URL anywhere in a commit message.
#   trello-notify.sh cards <git-range> [version]   one webhook call per distinct card in the range
#   trello-notify.sh finalize <version>            one webhook call that closes the version list
# Needs TRACT_URL and TRACT_TOKEN. Without them, or when a call fails, it only warns: a Trello outage
# must not fail a release.
set -u

mode=${1:?usage: trello-notify.sh cards <git-range> [version] | finalize <version>}

if [ -z "${TRACT_URL:-}" ] || [ -z "${TRACT_TOKEN:-}" ]; then
  echo "::notice::TRACT_URL/TRACT_TOKEN not set, Trello is not notified"
  exit 0
fi

post() {
  curl -sS -m 30 -o /dev/null -w "%{http_code}" -X POST \
    -H "Content-Type: application/json" \
    -H "X-Gitlab-Token: ${TRACT_TOKEN}" \
    -d "$1" "${TRACT_URL}"
}

notify() {
  status=$(post "$1") || status=curl-failed
  if [ "$status" != "200" ]; then
    echo "::warning::tract webhook answered ${status} for $1"
  fi
}

case "$mode" in
  cards)
    range=${2:?git range}
    version=${3:-}
    base=${range%%..*}
    if [ "$base" != "$range" ] && ! git rev-parse --verify -q "${base}^{commit}" >/dev/null; then
      range=HEAD~1..HEAD
    fi
    for card in $(git log --format=%B "$range" | grep -oE 'trello\.com/c/[A-Za-z0-9]{8}' | cut -d/ -f3 | sort -u); do
      notify "{\"phase\":\"card\",\"card\":\"${card}\",\"version\":\"${version}\"}"
    done
    ;;
  finalize)
    notify "{\"phase\":\"finalize\",\"version\":\"${2:?version}\"}"
    ;;
  *)
    echo "unknown mode: $mode" >&2
    exit 2
    ;;
esac
