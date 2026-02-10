#!/bin/bash

sourcePath=$(dirname $(dirname $(dirname $(readlink -f $0))))

debug_meta(){
	ARCHES=("amd64")
}

meta_tools(){
	ARCH=$(arch)
    ARCHES=("amd64" "arm64")
	IMAGE_PREFIX="acr.aishu.cn/ict/applicationrunner-"
}

meta_containe_tag() {
	prefix="$1-*"
	if git_status="$(git status --porcelain 2>/dev/null)" && [[ -z "${git_status}" ]]; then
		GIT_TREE_STATE="clean"
	else
		GIT_TREE_STATE="dirty"
	fi
	
	local git_commit="$(git rev-parse HEAD 2>/dev/null)"

	local git_version="$(git describe --tags --abbrev=0 --match=${prefix} 2>/dev/null)"
	local dash_inversion="$(echo "${git_version}" | sed "s/[^-]//g")"
	if [[ "${dash_inversion}" == "---" ]]; then
		git_version="$(echo "${git_version}" | sed "s/-\([0-9]\{1,\}\)-g\([0-9a-f]\{7\}\)$/.\1+\2/")"
	fi
	if [[ "${GIT_TREE_STATE}" == "dirty" ]] && [ ! -z ${git_version} ];  then
		git_version+="-dirty"
	fi
	echo ${git_version#$1-}
}

meta_get_version_vars() {

	GIT_COMMIT="$(git rev-parse HEAD 2>/dev/null)"

	if git_status="$(git status --porcelain 2>/dev/null)" && [[ -z "${git_status}" ]]; then
		GIT_TREE_STATE="clean"
	else
		GIT_TREE_STATE="dirty"
	fi

	GIT_VERSION="$(git describe --tags --match='v*' 2>/dev/null)"
	DASH_IN_VERSION="$(echo "${GIT_VERSION}" | sed "s/[^-]//g")"
	if [[ "${DASH_IN_VERSION}" == "---" ]]; then
		GIT_VERSION="$(echo "${GIT_VERSION}" | sed "s/-\([0-9]\{1,\}\)-g\([0-9a-f]\{7\}\)$/.\1+\2/")"
	fi
	if [[ "${GIT_TREE_STATE}" == "dirty" ]]; then
		GIT_VERSION+="-dirty"
	fi

	VERSION="${GIT_VERSION#v}"

    GIT_TAG=`echo ${VERSION} | awk -F "-" '{print $1}'`

	CURRENT_TAG="$(git describe --tags --abbrev=0 --match='v*' 2>/dev/null)"

	DOCKER_TAG="${DOCKER_TAG:-${GIT_VERSION#v}}"
	DOCKER_TAG="${DOCKER_TAG/+/_}"

	DOCKER_REGISTRY="${DOCKER_REGISTRY:-acr.aishu.cn}"
	DOCKER_REPOSITORY="${DOCKER_REPOSITORY:-as/deploy/tools}"
}

meta_go_ldflags() {
  LDFLAGS="-w -s"
  LDFLAGS+=" -X taskrunner/trait.GitCommit=${GIT_COMMIT}"
  LDFLAGS+=" -X taskrunner/trait.GitVersion=${GIT_VERSION}"
  LDFLAGS+=" -X taskrunner/trait.GitTreeState=${GIT_TREE_STATE}"
  LDFLAGS+=" -X taskrunner/trait.BuildDate=$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
}

meta_env() {
  workDir=$(dirname $(dirname $(dirname $(readlink -f $0))))
#   export GOPROXY=https://goproxy.aishu.cn,direct
}

meta_git_branchname() {
        GIT_BRANCHNAME=`git rev-parse --abbrev-ref HEAD`
        if [[ "${GIT_BRANCHNAME}"x = "HEAD"x ]]
        then
           GIT_BRANCHNAME=${BUILD_SOURCEBRANCHNAME}
        fi
}

set -e

meta_env
cd ${workDir}
meta_get_version_vars
meta_go_ldflags
meta_tools
meta_git_branchname
set +e
