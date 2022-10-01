#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

make generate

changes=$(git status --porcelain)
if [ -n "${changes}" ]; then
  echo "ERROR: files are not to date; please run: dev/scripts/generate.sh"
  echo "changed files:"
  printf "%s" "${changes}\n"
  echo "git diff:"
  git --no-pager diff
  exit 1
fi
