# This is where the final release artifacts are created locally
readonly RELEASE_STAGE="${LOCAL_OUTPUT_ROOT}/release-stage"
readonly RELEASE_TARS="${LOCAL_OUTPUT_ROOT}/release-tars"
readonly RELEASE_IMAGES="${LOCAL_OUTPUT_ROOT}/release-images"


function proton::release::package_tarballs() {
  # Clean out any old releases
  rm -rf "${RELEASE_STAGE}" "${RELEASE_TARS}" "${RELEASE_IMAGES}"
  mkdir -p "${RELEASE_TARS}"
  # proton::release::package_src_tarball &
  proton::release::package_bin_tarball &
  # proton::release::package_client_tarballs &
  # proton::release::package_proton_manifests_tarball &
  proton::util::wait-for-jobs || { proton::log::error "previous tarball phase failed"; return 1; }
}

# Package the source code we built, for compliance/licensing/audit/yadda.
function proton::release::package_src_tarball() {
  local -r src_tarball="${RELEASE_TARS}/proton-cli-src.tar.gz"
  proton::log::status "Building tarball: src"
  if [[ "${PROTON_GIT_TREE_STATE-}" = 'clean' ]]; then
    git archive -o "${src_tarball}" HEAD
  else
    find "${PROTON_ROOT}" -mindepth 1 -maxdepth 1 \
      ! \( \
        \( -path "${PROTON_ROOT}"/_\*       -o \
           -path "${PROTON_ROOT}"/.git\*    -o \
           -path "${PROTON_ROOT}"/.config\* -o \
           -path "${PROTON_ROOT}"/.gsutil\*    \
        \) -prune \
      \) -print0 \
    | "${TAR}" czf "${src_tarball}" --transform "s|${PROTON_ROOT#/*}|proton-cli|" --null -T -
  fi
}

# Package up all of the cross compiled binaries. Over time this should grow into
# a full SDK
function proton::release::package_bin_tarball() {
  # Find all of the built bin binaries
  local long_platforms=("${PROTON_OUTPUT_BINPATH}"/*/*)
  if [[ -n ${PROTON_BUILD_PLATFORMS-} ]]; then
    read -ra long_platforms <<< "${PROTON_BUILD_PLATFORMS}"
  fi

  for platform_long in "${long_platforms[@]}"; do
    local platform
    local platform_tag
    platform=${platform_long##${PROTON_OUTPUT_BINPATH}/} # Strip LOCAL_OUTPUT_BINPATH
    platform_tag=${platform/\//-} # Replace a "/" for a "-"
    proton::log::status "Starting tarball: bin $platform_tag"

    (
      # create staging directory
      rm -rf "${RELEASE_STAGE}/${platform}/proton-cli/bin"
      mkdir -p "${RELEASE_STAGE}/${platform}/proton-cli/bin"
      find "${PROTON_OUTPUT_BINPATH}/${platform}" -type f -exec rsync -pc {} "${RELEASE_STAGE}/${platform}/proton-cli/bin" \;

      local package_name="${RELEASE_TARS}/proton-cli-${platform_tag}.tar.gz"
      proton::release::create_tarball "${package_name}" "${RELEASE_STAGE}/${platform}"
    ) &
  done

  proton::log::status "Waiting on tarballs"
  proton::util::wait-for-jobs || { proton::log::error "client tarball creation failed"; exit 1; }
}

# Build a release tarball.  $1 is the output tar name.  $2 is the base directory
# of the files to be packaged. This assumes that ${2}/kubernetes is what is
# being packaged.
function proton::release::create_tarball() {
  proton::build::ensure_tar

  local tarfile=$1
  local stagingdir=$2

  "${TAR}" czf "${tarfile}" -C "${stagingdir}" proton-cli --owner=0 --group=0
}
