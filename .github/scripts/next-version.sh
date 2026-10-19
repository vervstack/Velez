#!/bin/sh
# Prints the next version tag, counted from the highest vMAJOR.MINOR.PATCH tag (a -rc suffix is
# ignored when comparing):
#   rc       vMAJOR.MINOR.PATCH+1-rc   the candidate every master push publishes
#   release  vMAJOR.MINOR+1.0          the manual release promoted from a candidate
set -eu

kind=${1:?usage: next-version.sh rc|release}

latest=$(git tag --list 'v[0-9]*.[0-9]*.[0-9]*' \
  | sed -e 's/^v//' -e 's/-rc$//' \
  | sort -t. -k1,1n -k2,2n -k3,3n \
  | tail -n1)
latest=${latest:-0.0.0}

major=${latest%%.*}
rest=${latest#*.}
minor=${rest%%.*}
patch=${rest#*.}

case "$kind" in
  rc) echo "v${major}.${minor}.$((patch + 1))-rc" ;;
  release) echo "v${major}.$((minor + 1)).0" ;;
  *) echo "unknown kind: $kind" >&2; exit 2 ;;
esac
