#!/usr/bin/env bash

# https://github.com/helm/helm/issues/4482#issuecomment-452013778

set -o errexit
set -o nounset
set -o pipefail

PKG_ROOT=$(realpath "$(dirname ${BASH_SOURCE[0]})/..")

cd $PKG_ROOT/helm

mkdir -p tmpcharts
trap "rmdir tmpcharts" EXIT

# sign charts using user.signingkey if git commit signing is enabled
# shellcheck disable=SC2046
helm package $(git config --get commit.gpgsign | grep -q true && echo "--sign") \
  --key $(git config --get user.signingkey) --keyring "$HOME/.gnupg/secring.gpg" \
  -d tmpcharts powergrid
helm repo index --url "https://sportshead.github.io/powergrid" --merge charts/index.yaml tmpcharts

mv tmpcharts/* charts/
