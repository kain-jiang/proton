readonly PROTON_BUILD_IMAGE_DEPENDENCIES=(
  build/dev/Dockerfile.build
  build/update-build-image-withbuildx
  cmd/download-proton-cli-web
  cmd/run-pipeline
  go.sum
)

function proton::build-image::tag {
  git log -1 --pretty=%h --abbrev=8 "${PROTON_BUILD_IMAGE_DEPENDENCIES[@]}"
}
