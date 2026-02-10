readonly PROTON_GO_PACKAGE=devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3
readonly PROTON_GOPATH="${PROTON_OUTPUT}/go"

readonly PROTON_SUPPORTED_PLATFORMS=(
  linux/amd64
  linux/arm64
)

declare -a PROTON_PLATFORMS
function proton::golang::setup_platforms {
  if [[ -n "${PROTON_BUILD_PLATFORMS:-}" ]]; then
    local -a platforms
    IFS=" " read -ra platforms <<< "${PROTON_BUILD_PLATFORMS}"
    PROTON_PLATFORMS=("${platforms[@]}")
  else
    PROTON_PLATFORMS=("${PROTON_SUPPORTED_PLATFORMS[@]}")
  fi
  readonly PROTON_PLATFORMS
}
proton::golang::setup_platforms

# The set of client targets that we are building for all platforms
readonly PROTON_TARGETS=(
  cmd/proton-cli
)
readonly PROTON_BINARIES=("${PROTON_TARGETS[@]##*/}")

# Asks golang what it thinks the host platform is. The go tool chain does some
# slightly different things when the target platform matches the host platform.
function proton::golang::host_platform {
  echo "$(go env GOHOSTOS)/$(go env GOHOSTARCH)"
}

# Takes the platform name ($1) and sets the appropriate golang env variables
# for that platform.
function proton::golang::set_platform_envs {
  [[ -n ${1-} ]] || {
    proton::log::error_exit "!!! Internal error. No platform set in proton::golang::set_platform_envs"
  }

  export GOOS=${platform%/*}
  export GOARCH=${platform##*/}

  # Do not set CC when building natively on a platform, only if cross-compiling
  if [[ $(proton::golang::host_platform) != "$platform" ]]; then
    # Dynamic CGO linking for other server architectures than host architecture goes here
    # If you want to include support for more server platforms than these, add arch-specific gcc names here
    case "${platform}" in
      "linux/amd64")
        export CGO_ENABLED=1
        export CC=${PROTON_LINUX_AMD64_CC:-x86_64-linux-gnu-gcc}
        ;;
      "linux/arm")
        export CGO_ENABLED=1
        export CC=${PROTON_LINUX_ARM_CC:-arm-linux-gnueabihf-gcc}
        ;;
      "linux/arm64")
        export CGO_ENABLED=1
        export CC=${PROTON_LINUX_ARM64_CC:-aarch64-linux-gnu-gcc}
        ;;
      "linux/ppc64le")
        export CGO_ENABLED=1
        export CC=${PROTON_LINUX_PPC64LE_CC:-powerpc64le-linux-gnu-gcc}
        ;;
      "linux/s390x")
        export CGO_ENABLED=1
        export CC=${PROTON_LINUX_S390X_CC:-s390x-linux-gnu-gcc}
        ;;
    esac
  fi

  # if CC is defined for platform then always enable it
  ccenv=$(echo "$platform" | awk -F/ '{print "PROTON_" toupper($1) "_" toupper($2) "_CC"}')
  if [ -n "${!ccenv-}" ]; then
    export CGO_ENABLED=1
    export CC="${!ccenv}"
  fi
}

function proton::golang::unset_platform_envs() {
  unset GOOS
  unset GOARCH
  unset GOROOT
  unset CGO_ENABLED
  unset CC
}

# Ensure the go tool exists and is a viable version.
function proton::golang::verify_go_version {
  if [[ -z "$(command -v go)" ]]; then
    proton::log::usage_from_stdin <<EOF
Can't find 'go' in PATH, please fix and retry.
See http://golang.org/doc/install for installation instructions.
EOF
    return 2
  fi

  local go_version
  IFS=" " read -ra go_version <<< "$(GOFLAGS='' go version)"
  local minimum_go_version
  minimum_go_version=go1.18.0
  if [[ "${minimum_go_version}" != $(echo -e "${minimum_go_version}\n${go_version[2]}" | sort -s -t. -k 1,1 -k 2,2n -k 3,3n | head -n1) && "${go_version[2]}" != "devel" ]]; then
    proton::log::usage_from_stdin <<EOF
Detected go version: ${go_version[*]}.
Proton requires ${minimum_go_version} or greater.
Please install ${minimum_go_version} or later.
EOF
    return 2
  fi
}

# proton::golang::setup_env will check that the `go` commands is available in
# ${PATH}. It will also check that the Go version is good enough for the
# Kubernetes build.
#
# Outputs:
#   env-var GOROOT
#   env-var GOBIN is unset (we want binaries in a predictable place)
#   env-var GO15VENDOREXPERIMENT=1
function proton::golang::setup_env {
  proton::golang::verify_go_version

  # Set GOROOT so binaries that parse code can work properly.
  GOROOT=$(go env GOROOT)
  export GOROOT

  # Unset GOBIN in case it already exists in the current session.
  unset GOBIN
}

# This will take binaries from $GOPATH/bin and copy them to the appropriate
# place in ${KUBE_OUTPUT_BINDIR}
#
# Ideally this wouldn't be necessary and we could just set GOBIN to
# KUBE_OUTPUT_BINDIR but that won't work in the face of cross compilation.  'go
# install' will place binaries that match the host platform directly in $GOBIN
# while placing cross compiled binaries into `platform_arch` subdirs.  This
# complicates pretty much everything else we do around packaging and such.
function proton::golang::place_bins {
  local host_platform
  host_platform=$(proton::golang::host_platform)

  V=2 proton::log::status "Placing binaries"

  local platform
  for platform in "${PROTON_PLATFORMS[@]}"; do
    # The substitution on platform_src below will replace all slashes with
    # underscores.  It'll transform darwin/amd64 -> darwin_amd64.
    local platform_src="/${platform//\//_}"
    if [[ "${platform}" == "${host_platform}" ]]; then
      platform_src=""
      rm -f "${THIS_PLATFORM_BIN}"
      ln -s "${PROTON_OUTPUT_BINPATH}/${platform}" "${THIS_PLATFORM_BIN}"
    fi

    local full_binpath_src="${PROTON_GOPATH}/bin${platform_src}"
    if [[ -d "${full_binpath_src}" ]]; then
      mkdir -p "${PROTON_OUTPUT_BINPATH}/${platform}"
      find "${full_binpath_src}" -maxdepth 1 -type f -exec \
        rsync -pc {} "${PROTON_OUTPUT_BINPATH}/${platform}" \;
    fi
  done
}

function proton::golang::binaries_from_targets {
  local target
  for target; do
    echo "${PROTON_GO_PACKAGE}/${target}"
  done
}

function proton::golang::build_some_binaries {
  for binary; do
    go build "${build_args[@]}" -o "${LOCAL_OUTPUT_ROOT}/local/bin/${platform}/${binary##*/}" "${binary}"
  done
}

function proton::golang::build_binaries_for_platform {
  local platform=$1

  V=2 proton::log::info "Env for ${platform}: GOOS=${GOOS-} GOARCH=${GOARCH-} GOROOT=${GOROOT-} GOPATH=${GOPATH-} CGO_ENABLED=${CGO_ENABLED-} CC=${CC-}"

  # except -o, it's set by binary
  local -a build_args
  build_args=(
    -ldflags "${goldflags:-}"
  )
  V=1 proton::log::info "> build CGO_ENABLED=0: ${binaries[*]}"
  CGO_ENABLED=0 proton::golang::build_some_binaries "${binaries[@]}"
}

function proton::golang::build_binaries {
  # Create a sub-shell so that we don't pollute the outer environment
  (
    # Check for `go` binary
    proton::golang::setup_env
    V=2 proton::log::info "GO version: $(GOFLAGS='' go version)"

    local host_platform
    host_platform=$(proton::golang::host_platform)

    local -a targets=()
    local arg
    for arg; do
      targets+=("${arg}")
    done

    if [[ ${#targets[@]} -eq 0 ]]; then
      targets=("${PROTON_TARGETS[@]}")
    fi

    local goldflags
    goldflags="${GOLDFLAGS=-s -w} $(proton::log::ldflags) $(proton::version::ldflags)"

    local -a platforms
    IFS=" " read -ra platforms <<< "${PROTON_BUILD_PLATFORMS:-}"
    if [[ ${#platforms[@]} -eq 0 ]]; then
      platforms=("${host_platform}")
    fi

    local -a binaries
    while IFS="" read -r binary; do binaries+=("$binary"); done < <(proton::golang::binaries_from_targets "${targets[@]}")

    for platform in "${platforms[@]}"; do
      proton::log::status "Building go targets for ${platform}"
      (
        proton::golang::set_platform_envs "${platform}"
        proton::golang::build_binaries_for_platform "${platform}"
      )
    done
  )
}
