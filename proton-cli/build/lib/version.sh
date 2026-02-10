function proton::version::get_version_vars {
  # 显式指定长度避免平台不同生成的 git commit id 长度不同
  if PROTON_VERSION=$(git describe --tags --match='v*' --abbrev=7 2>/dev/null); then
    # remove prefix "v" -> 1.2.3-alpha.0-4-g09078d4
    PROTON_VERSION="${PROTON_VERSION/v/}"
    DASH_IN_VERSION=$(echo "${PROTON_VERSION}" | sed 's/[^-]//g')
    if [[ "${DASH_IN_VERSION}" == "---" ]]; then
      # 1.2.3-alpha.0-4-g09078d4 -> 1.2.3.4-alpha.0-4+g09078d4
      PROTON_VERSION=$(echo "${PROTON_VERSION}" | sed 's/-\([0-9]\{1,\}\)-g\([0-9a-f]\{7\}\)$/.\1+\2/')
    elif [[ "${DASH_IN_VERSION}" == "--" ]]; then
      # 1.2.3-alpha.0-g09078d4 -> 1.2.3.4-alpha.0+g09078d4
      PROTON_VERSION=$(echo "${PROTON_VERSION}" | sed 's/-g\([0-9a-f]\{7\}\)$/+\1/')
    fi
  fi
  # Azure DevOps Pipeline Build Number
  if [[ -n "${BUILD_BUILDNUMBER-}" ]]; then
    PROTON_VERSION="${PROTON_VERSION}.${BUILD_BUILDNUMBER}"
  fi

  PROTON_BUILD_DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ)

  PROTON_GIT_COMMIT=$(git rev-parse HEAD)

  if git_status=$(git status --porcelain 2>/dev/null) && [[ -z "${git_status}" ]]; then
    PROTON_GIT_TREE_STATE="clean"
  else
    PROTON_GIT_TREE_STATE="dirty"
    PROTON_VERSION+="-dirty"
  fi

  # no change, just for compatibility
  # 1.2.3-alpha.0.4
  PROTON_SEMANTIC_VERSION=$(echo ${PROTON_VERSION} | sed 's/\([0-9]\{1,\}\)\.\([0-9]\{1,\}\)\.\([0-9]\{1,\}\)\.\([0-9]\{1,\}\)/\1.\2.\3-\4/')

  # no change, just for compatibility
  # 1.2.3-alpha.0.4
  PROTON_CONTAINER_IMAGE_TAG="${PROTON_VERSION/+/_}"
}

function proton::version::ldflags {
  proton::version::get_version_vars

  local -a ldflags
  function add_ldflag() {
    local key=$1
    local val=$2
    ldflags+=(
      "-X '${PROTON_GO_PACKAGE}/pkg/version.${key}=${val}'"
    )
  }

  add_ldflag "gitVersion"   "${PROTON_VERSION}"
  add_ldflag "gitCommit"    "${PROTON_GIT_COMMIT}"
  add_ldflag "gitTreeState" "${PROTON_GIT_TREE_STATE}"
  add_ldflag "buildDate"    "${PROTON_BUILD_DATE}"

  echo "${ldflags[*]}"
}
