#!/usr/bin/env sh
set -eu

secret_file=/home/desk/secred-github

if [ ! -r "$secret_file" ]; then
  echo "GitHub token file is not readable: $secret_file" >&2
  exit 1
fi

IFS= read -r token < "$secret_file" || [ -n "${token:-}" ]

case "$token" in
  ghp_*|github_pat_*|gho_*|ghs_*|ghu_*|ghr_*) ;;
  *)
    echo "GitHub token file does not contain a supported token format" >&2
    exit 1
    ;;
esac

case "$token" in
  *[!A-Za-z0-9_]*)
    echo "GitHub token contains unexpected characters" >&2
    exit 1
    ;;
esac

printf '{"Authorization":"Bearer %s"}\n' "$token"
