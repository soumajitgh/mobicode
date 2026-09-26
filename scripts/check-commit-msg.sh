#!/bin/sh
set -eu

subject=$(sed -n '1p' "$1")

case "$subject" in
	Merge\ *|Revert\ \"*|fixup\!\ *|squash\!\ *) exit 0 ;;
esac

if printf '%s\n' "$subject" | grep -Eq '^(feat|fix|docs|style|refactor|perf|test|build|ci|chore|revert)(\([a-z0-9][a-z0-9._/-]*\))?!?: [^[:space:]].*$'; then
	exit 0
fi

printf 'Invalid commit subject: %s\n' "$subject" >&2
printf 'Use: type(scope): description (for example, feat(server): add hot reload)\n' >&2
exit 1
