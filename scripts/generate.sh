#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd -- "${SCRIPT_DIR}/.." && pwd)"
PREPARED_SPEC="${ROOT_DIR}/openapi/public.swagger.json"
SERVICES_DIR="${ROOT_DIR}/openapi/services"
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "${TMP_DIR}"' EXIT

cd "${ROOT_DIR}"

ALL_SERVICES=(feeds videos images)

require_command() {
	local command="$1"
	if ! command -v "${command}" >/dev/null 2>&1; then
		echo "${command} is required" >&2
		exit 1
	fi
}

usage() {
	cat <<'EOF'
Usage:
  ./scripts/generate.sh [--service feeds|videos|images] [--spec /path/to/openapi.json]

Defaults:
  - generates all services when --service is omitted
	- uses openapi/services/<service>.swagger.json when --spec is omitted
EOF
}

validate_service() {
	local service="$1"
	case "${service}" in
		feeds|videos|images) ;;
		*)
			echo "unsupported service: ${service}" >&2
			echo "supported services: feeds, videos, images" >&2
			exit 1
			;;
	esac
}

service_arg=""
spec_arg=""

while [[ $# -gt 0 ]]; do
	case "$1" in
		--service)
			[[ $# -ge 2 ]] || { echo "--service requires a value" >&2; exit 1; }
			service_arg="$2"
			shift 2
			;;
		--spec)
			[[ $# -ge 2 ]] || { echo "--spec requires a value" >&2; exit 1; }
			spec_arg="$2"
			shift 2
			;;
		-h|--help|help)
			usage
			exit 0
			;;
		*)
			echo "unknown argument: $1" >&2
			usage
			exit 1
			;;
	esac
done

if [[ -n "${service_arg}" ]]; then
	validate_service "${service_arg}"
fi

require_command java
require_command npx
require_command rsync
require_command jq

services=("${ALL_SERVICES[@]}")
if [[ -n "${service_arg}" ]]; then
	services=("${service_arg}")
fi

mkdir -p "${ROOT_DIR}/sdk" "${SERVICES_DIR}"

if [[ -n "${spec_arg}" ]]; then
	cp "${spec_arg}" "${PREPARED_SPEC}"
	for service in "${services[@]}"; do
		service_spec="${SERVICES_DIR}/${service}.swagger.json"
		bash "${ROOT_DIR}/../api/sdk/openapi/split-spec.sh" "${PREPARED_SPEC}" "${service}" "${service_spec}"
	done
fi

for service in "${services[@]}"; do
	service_spec="${SERVICES_DIR}/${service}.swagger.json"
	service_tmp="${TMP_DIR}/out-${service}"
	output_dir="${ROOT_DIR}/sdk/${service}"

	if [[ ! -f "${service_spec}" ]]; then
		echo "service spec not found at ${service_spec}" >&2
		echo "run with --spec /path/to/openapi.json to refresh local service specs" >&2
		exit 1
	fi

	npx -y @openapitools/openapi-generator-cli generate \
		-g go \
		-i "${service_spec}" \
		-o "${service_tmp}" \
		--global-property apiDocs=false,modelDocs=false,apiTests=false,modelTests=false \
		--additional-properties "packageName=${service},packageVersion=2.0.0,generateInterfaces=true,enumClassPrefix=true,withGoMod=false,isGoSubmodule=true,hideGenerationTimestamp=true,disallowAdditionalPropertiesIfNotPresent=false"

	rm -rf "${service_tmp}/.openapi-generator" "${service_tmp}/api"
	rm -f \
		"${service_tmp}/.travis.yml" \
		"${service_tmp}/git_push.sh" \
		"${service_tmp}/README.md" \
		"${service_tmp}/.gitignore" \
		"${service_tmp}/.openapi-generator-ignore"

	mkdir -p "${output_dir}"
	rsync -a --delete \
		--exclude '.git' \
		--exclude '.github' \
		"${service_tmp}/" "${output_dir}/"
done

echo "Generated Go SDK services under ${ROOT_DIR}/sdk"
