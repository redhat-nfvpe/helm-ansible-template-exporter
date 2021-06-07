#!/usr/bin/env bash

#
# helmExport.sh
#
# Given an input Helm chart directory, an output workspace directory, and an output role name, helmExport.sh makes a
# best effort conversion of the input Helm chart to a corresponding Ansible Role.
#

CONTAINER_CHART_PATH="/chart"
CONTAINER_WORKSPACE_PATH="/workspace"

HELM_EXPORT_EXECUTABLE="helmExport"
HELM_EXPORT_BIN="/usr/local/helmExport/bin"
HELM_EXPORT_PATH="${HELM_EXPORT_BIN}/${HELM_EXPORT_EXECUTABLE}"

export REQUIRED_VARS=('HELM_CHART_PATH' 'OUTPUT_PATH' 'EMITTED_ROLE')
export REQUIRED_VARS_ERROR_MESSAGES=(
	'HELM_CHART_PATH is invalid or not given. Use the -c option to provide path to the input helm chart.'
	'OUTPUT_PATH is required. Use the -w option to specify the directory containing the output workspace.'
	'EMITTED_ROLE is required. Use the -r option to specify the output Ansible Role name.'
)

# outputs a welcome banner to stdout
output_welcome_banner() {
  echo "Helm Ansible Template Exporter}"
}

# usage contains usage information
usage() {
	read -d '' usage_prompt <<- EOF
  Usage: $0 -c HELM_CHART_PATH -w OUTPUT_PATH -r EMITTED_ROLE

  Export a helm chart as an Ansible Role.  Note, not all cases are covered, and this tool is considered best effort.

  Options (required)
    -c: set the directory containing the input helm chart
    -w: set the output working directory where the Ansible Playbook Role is emitted
    -r: set the name of the role for the generated Ansible Playbook Role
	EOF

	echo -e "$usage_prompt"
}

# usage_error emits the usage message and then exits with exit code 1.
usage_error() {
	usage
	exit 1
}

# check_required_vars checks that the required bash arguments have been provided.
check_required_vars() {
	local var_missing=false

	for index in "${!REQUIRED_VARS[@]}"; do
		var=${REQUIRED_VARS[$index]}
		if [[ -z ${!var} ]]; then
			error_message=${REQUIRED_VARS_ERROR_MESSAGES[$index]}
			echo "$0: error: $error_message" 1>&2
			var_missing=true
		fi
	done

	if $var_missing; then
		echo ""
		usage_error
	fi
}

# parse_cli_args parses the input CLI arguments.
parse_cli_args() {
  # Parse args beginning with -
  while [[ "$1" == -* ]]; do
    echo "$1 $2"
      case "$1" in
        -h|--help|-\?) usage; exit 0;;
        -c) if (($# > 1)); then
              export HELM_CHART_PATH="$2"
              shift 2
            else
              echo "-c requires an argument" 1>&2
              exit 1
            fi ;;
        -w) if (($# > 1)); then
              export OUTPUT_PATH="$2"; shift 2
            else
              echo "-w requires an argument" 1>&2
              exit 1
            fi ;;
        -r) if (($# > 1)); then
              export EMITTED_ROLE="$2"; shift 2
            else
              echo "-r requires an argument" 1>&2
              exit 1
            fi ;;
        --) shift; break;;
        -*) echo "invalid option: $1" 1>&2; usage_error;;
      esac
  done
}

# check_output_dir ensures that there isn't a generated role already in place in the output working directory.
check_output_dir() {
  if [ -d "${OUTPUT_PATH}/${EMITTED_ROLE}" ]
  then
    echo "output directory \"${OUTPUT_PATH}/${EMITTED_ROLE}\" already has a generated role;  delete or change the output directory before proceeding"
    exit 2
  else
    echo "Ansible Role \"${EMITTED_ROLE}\" will be generated in \"${OUTPUT_PATH}\""
  fi
}

output_welcome_banner
parse_cli_args "$@"
check_required_vars
mkdir -p "${OUTPUT_PATH}"
check_output_dir

# runs the conversion utility
docker run -v "${HELM_CHART_PATH}":"${CONTAINER_CHART_PATH}":Z \
    -v "${OUTPUT_PATH}":"${CONTAINER_WORKSPACE_PATH}":Z \
    helm-export:v1.0.0 "${HELM_EXPORT_PATH}" export "${EMITTED_ROLE}" --helm-chart="${CONTAINER_CHART_PATH}" --workspace="${CONTAINER_WORKSPACE_PATH}" 2> "${OUTPUT_PATH}/${EMITTED_ROLE}-conversion-log.txt"
