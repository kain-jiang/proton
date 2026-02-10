# Control verbosity of the script output and logging.
PROTON_VERBOSE="${PROTON_VERBOSE:-2}"

# Log an error but keep going.  Don't dump the stack or exit.
function proton::log::error {
  timestamp=$(date +"[%m%d %H:%M:%S]")
  echo "!!! ${timestamp} ${1-}" >&2
  shift
  for message; do
    echo "    ${message}" >&2
  done
}

# Print an usage message to stderr.  The arguments are printed directly.
function proton::log::usage {
  echo >&2
  local message
  for message; do
    echo "${message}" >&2
  done
  echo >&2
}

function proton::log::usage_from_stdin {
  local -a messages
  while read -r line; do
    messages+=("${line}")
  done

  proton::log::usage "${messages[@]}"
}

function proton::log::info {
  local V="${V:-0}"
  if [[ ${PROTON_VERBOSE} < ${V} ]]; then
    return
  fi

  for message; do
    echo "${message}"
  done
}

function proton::log::status {
  local V="${V:-0}"
  if [[ ${PROTON_VERBOSE} < ${V} ]]; then
    return
  fi

  timestamp=$(date +"[%m%d %H:%M:%S]")
  echo "+++ ${timestamp} ${1}"
  shift
  for message; do
    echo "    ${message}"
  done
}

# just for debugging
function proton::log::execute {
  local V="${V:-3}"
  if [[ ${PROTON_VERBOSE} < ${V} ]]; then
    return
  fi

  timestamp=$(date +"[%m%d %H:%M:%S]")
  echo "${timestamp} execute ${1}"
  shift
  for message; do
    echo "    args: ${message}"
  done
}

function proton::log::get_log_vars {
  if [[ $(git describe --tags) =~ .*"-".* ]]; then
    PROTON_DEFAULT_LOG_LEVEL="debug"
  else
    PROTON_DEFAULT_LOG_LEVEL="info"
  fi
}

function proton::log::ldflags {
  proton::log::get_log_vars

  local -a ldflags
  function add_ldflag() {
    local key=$1
    local val=$2
    ldflags+=(
      "-X '${PROTON_GO_PACKAGE}/${key}=${val}'"
    )
  }

  add_ldflag "pkg/core/global.LoggerLevel"  "${PROTON_DEFAULT_LOG_LEVEL}"

  echo "${ldflags[*]}"
}
