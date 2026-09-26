#!/bin/sh
set -eu

repo_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$repo_root"

go_cmd=${GO:-go}
pnpm_cmd=${PNPM:-pnpm}
make_cmd=${MAKE:-make}

for cmd in "$go_cmd" node "$pnpm_cmd" "$make_cmd"; do
	if ! command -v "$cmd" >/dev/null 2>&1; then
		printf 'Missing required command: %s\n' "$cmd" >&2
		exit 1
	fi
done

if ! node -e 'process.exit(Number(process.versions.node.split(".")[0]) >= 22 ? 0 : 1)'; then
	printf 'Node.js 22 or newer is required.\n' >&2
	exit 1
fi

if [ ! -f .env ]; then
	cp .env.example .env
	secret=$(openssl rand -hex 32)
	sed -i.bak "s/^MOBICODE_SERVER_SECRET_TOKEN=.*/MOBICODE_SERVER_SECRET_TOKEN=$secret/" .env
	rm .env.bak
	printf 'Created .env from .env.example\n'
fi

printf 'Downloading Go dependencies...\n'
"$go_cmd" mod download

printf 'Installing root web dependencies...\n'
"$pnpm_cmd" install --frozen-lockfile

if git rev-parse --git-dir >/dev/null 2>&1; then
	printf 'Installing commit message hook...\n'
	"$pnpm_cmd" exec lefthook install
fi

printf 'Installing mobile dependencies...\n'
"$make_cmd" mobile/install GO="$go_cmd" PNPM="$pnpm_cmd"

printf 'Installing website dependencies...\n'
"$make_cmd" website/install GO="$go_cmd" PNPM="$pnpm_cmd"

printf 'Building server and browser assets...\n'
"$make_cmd" server/build GO="$go_cmd" PNPM="$pnpm_cmd"

printf 'Setup complete. Run make server/dev, make mobile/start, or make website/start.\n'
