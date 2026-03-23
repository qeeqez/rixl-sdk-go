#!/usr/bin/env bash

set -euo pipefail

if [[ $# -ne 3 ]]; then
	echo "usage: $0 <input-spec> <service> <output-spec>" >&2
	exit 1
fi

INPUT_SPEC="$1"
SERVICE="$2"
OUTPUT_SPEC="$3"

if ! command -v jq >/dev/null 2>&1; then
	echo "jq is required to split the public SDK spec" >&2
	exit 1
fi

if [[ ! -f "${INPUT_SPEC}" ]]; then
	echo "input spec not found at ${INPUT_SPEC}" >&2
	exit 1
fi

case "${SERVICE}" in
	feeds)
		TAG="Feeds"
		;;
	videos)
		TAG="Videos"
		;;
	images)
		TAG="Images"
		;;
	*)
		echo "unsupported service: ${SERVICE}" >&2
		echo "supported services: feeds, videos, images" >&2
		exit 1
		;;
esac

mkdir -p "$(dirname -- "${OUTPUT_SPEC}")"

jq --arg tag "${TAG}" '
	def definition_ref_name:
		select(type == "string")
		| select(startswith("#/definitions/"))
		| sub("^#/definitions/"; "");

	def collect_definition_refs:
		[
			..
			| objects
			| .["$ref"]? // empty
			| definition_ref_name
		] | unique;

	def expand_definition_refs($definitions; $pending; $seen):
		if ($pending | length) == 0 then
			$seen
		else
			($pending - $seen) as $next_batch
			| if ($next_batch | length) == 0 then
				$seen
			else
				(
					[
						$next_batch[]
						| $definitions[.]? // empty
						| collect_definition_refs[]
					] | unique
				) as $nested
				| expand_definition_refs(
					$definitions;
					($nested + $next_batch | unique);
					($seen + $next_batch | unique)
				)
			end
		end;

	(.definitions // {}) as $definitions
	| .tags = [(.tags // [])[] | select(.name == $tag)]
	| .paths |= with_entries(
		.value |= with_entries(select((.value.tags // []) | index($tag)))
		| select((.value | length) > 0)
	)
	| ((del(.definitions)) | collect_definition_refs) as $root_refs
	| expand_definition_refs($definitions; $root_refs; []) as $kept_definitions
	| .definitions = (
		$definitions
		| with_entries(select(.key as $name | $kept_definitions | index($name)))
	)
' "${INPUT_SPEC}" > "${OUTPUT_SPEC}"

echo "Prepared ${SERVICE} SDK spec at ${OUTPUT_SPEC}"
