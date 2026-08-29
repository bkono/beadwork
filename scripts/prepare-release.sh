#!/usr/bin/env bash
# Prepare a semver release from conventional commits since the latest tag.
# Writes CHANGELOG.md and cmd/bw/main.go, and sets GitHub Actions outputs:
#   release=true|false, version=x.y.z
set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$root"

github_output="${GITHUB_OUTPUT:-}"

current_tag="$(git describe --tags --abbrev=0 2>/dev/null || true)"
current_version="${current_tag#v}"
if [ -z "$current_version" ]; then
	current_version="0.0.0"
fi

next_tag="$(go tool svu next)"
next_version="${next_tag#v}"

if [ "$current_version" = "$next_version" ]; then
	echo "No release needed (still at v${current_version})"
	if [ -n "$github_output" ]; then
		echo "release=false" >>"$github_output"
	fi
	exit 0
fi

prev_ref="$current_tag"
if [ -z "$prev_ref" ]; then
	prev_ref="$(git rev-list --max-parents=0 HEAD)"
fi

mapfile -t commit_shas < <(git log "${prev_ref}..HEAD" --pretty=format:%H --no-merges)
if [ "${#commit_shas[@]}" -eq 0 ]; then
	echo "No commits found between ${prev_ref} and HEAD"
	exit 1
fi

commit_subject() {
	git log -1 --pretty=format:%s "$1"
}

commit_is_breaking() {
	local sha="$1"
	local subject body
	subject="$(git log -1 --pretty=format:%s "$sha")"
	body="$(git log -1 --pretty=format:%b "$sha")"
	[[ "$subject" =~ !: ]] && return 0
	grep -q '^BREAKING CHANGE:' <<<"$body" && return 0
	return 1
}

release_date="$(date -u +%Y-%m-%d)"
section_file="$(mktemp)"
trap 'rm -f "$section_file"' EXIT

{
	echo "## ${next_version} — ${release_date}"
	echo
} >"$section_file"

matches_prefix() {
	local subject="$1"
	shift
	local prefix
	for prefix in "$@"; do
		[[ "$subject" == "${prefix}"* ]] && return 0
	done
	return 1
}

append_group() {
	local title="$1"
	shift
	local -a prefixes=("$@")
	local found=false
	local sha subject

	for sha in "${commit_shas[@]}"; do
		subject="$(commit_subject "$sha")"
		if commit_is_breaking "$sha"; then
			continue
		fi
		if matches_prefix "$subject" "${prefixes[@]}"; then
			if [ "$found" = false ]; then
				echo "### ${title}" >>"$section_file"
				echo >>"$section_file"
				found=true
			fi
			echo "- ${subject}" >>"$section_file"
		fi
	done

	if [ "$found" = true ]; then
		echo >>"$section_file"
	fi
}

append_breaking_group() {
	local found=false
	local sha subject

	for sha in "${commit_shas[@]}"; do
		if ! commit_is_breaking "$sha"; then
			continue
		fi
		subject="$(commit_subject "$sha")"
		if [ "$found" = false ]; then
			echo "### Breaking Changes" >>"$section_file"
			echo >>"$section_file"
			found=true
		fi
		echo "- ${subject}" >>"$section_file"
	done
	if [ "$found" = true ]; then
		echo >>"$section_file"
	fi
}

append_breaking_group
append_group "Features" feat
append_group "Bug Fixes" fix
append_group "Performance" perf

other_found=false
for sha in "${commit_shas[@]}"; do
	subject="$(commit_subject "$sha")"
	if commit_is_breaking "$sha"; then
		continue
	fi
	if matches_prefix "$subject" feat fix perf chore ci docs test style build; then
		continue
	fi
	if [ "$other_found" = false ]; then
		echo "### Other" >>"$section_file"
		echo >>"$section_file"
		other_found=true
	fi
	echo "- ${subject}" >>"$section_file"
done
if [ "$other_found" = true ]; then
	echo >>"$section_file"
fi

if ! grep -q '^- ' "$section_file"; then
	echo "No changelog entries found between ${prev_ref} and HEAD"
	exit 1
fi

python3 - "$section_file" <<'PY'
import pathlib
import sys

section = pathlib.Path(sys.argv[1]).read_text()
changelog = pathlib.Path("CHANGELOG.md")
text = changelog.read_text()
marker = "## Unreleased\n"
if marker not in text:
    raise SystemExit("CHANGELOG.md is missing ## Unreleased section")

head, tail = text.split(marker, 1)
tail = tail.lstrip("\n")
changelog.write_text(head + marker + "\n" + section + tail)
PY

sed -i.bak "s/var version = \".*\"/var version = \"${next_version}\"/" cmd/bw/main.go
rm -f cmd/bw/main.go.bak

echo "Prepared release v${next_version}"
if [ -n "$github_output" ]; then
	echo "release=true" >>"$github_output"
	echo "version=${next_version}" >>"$github_output"
fi
