set -o errexit
set -o nounset
set -o pipefail

PROTON_ROOT="$(realpath $(dirname ${BASH_SOURCE[0]})/../..)"

PROTON_OUTPUT_SUBPATH="${PROTON_OUTPUT_SUBPATH:-_output/local}"
PROTON_OUTPUT="${PROTON_ROOT}/${PROTON_OUTPUT_SUBPATH}"
PROTON_OUTPUT_BINPATH="${PROTON_OUTPUT}/bin"

# This is a symlink to binaries for "this platform", e.g. build tools.
export THIS_PLATFORM_BIN="${PROTON_ROOT}/_output/bin"

source "${PROTON_ROOT}/build/lib/util.sh"
source "${PROTON_ROOT}/build/lib/log.sh"
source "${PROTON_ROOT}/build/lib/common.sh"

source "${PROTON_ROOT}/build/lib/build-image.sh"
source "${PROTON_ROOT}/build/lib/golang.sh"
source "${PROTON_ROOT}/build/lib/release.sh"
source "${PROTON_ROOT}/build/lib/version.sh"

source "${PROTON_ROOT}/build/lib/gocov.sh"
source "${PROTON_ROOT}/build/lib/golangci-lint.sh"
source "${PROTON_ROOT}/build/lib/gotestsum.sh"
