#!/usr/bin/bash

set -o errexit
set -o pipefail
set -o nounset

PKG_ROOT="$(dirname "${BASH_SOURCE[0]}")"

BIN_DIR="${PKG_ROOT}/bin"
REPO_DIR="$(realpath "${PKG_ROOT}/repos")"

function install_binaries {
  local -a binaries
  readarray -t binaries < <(ls "${BIN_DIR}")
  for src in "${binaries[@]}"; do
    dst="/usr/local/bin/$(basename "${src}")"
    echo "Installing ${dst}"
    install "${BIN_DIR}/${src}" "${dst}"
  done
}

function install_rpm_packages {
  # 1. create rpm repository config
  sed "s|BASEURL|${REPO_DIR}|" < "${REPO_DIR}/proton.repo.tmpl" > "/etc/yum.repos.d/proton.repo"

  # 2. execute dnf install
  local names=(
    containerd
    ecms
    haproxy
    kubeadm
    kubectl
    kubelet
    proton-cr
  )
  dnf install "${names[@]}" --repo=proton --allowerasing
}

function patch_for_distro {
  (
    if [[ ! -f /etc/os-release ]]; then
      echo >&2 "WARNING: /etc/os-release does't exist"
      return
    fi
    # 在子 shell 载入 /etc/os-release 避免污染当前 shell 环境
    source /etc/os-release
    case "${ID}-${VERSION_ID}" in
      kylin-V10)
        # 如果存在 libselinux.so.1，ecms 调用 systemctl 失败
        # systemctl: /usr/local/ecms/bin/libselinux.so.1: no version information available (required by /usr/lib/systemd/libsystemd-shared-243.so)
        # systemctl: /usr/local/ecms/bin/libselinux.so.1: no version information available (required by /usr/lib64/libmount.so.1)
        # systemctl: /usr/local/ecms/bin/libselinux.so.1: no version information available (required by /usr/lib64/libdevmapper.so.1.02)
        local target="/usr/local/ecms/bin/libselinux.so.1"
        if [[ -f "${target}" ]]; then
          rm "${target}"
        fi
        ;;

      *)
        echo "Not any patches for ${ID}-${VERSION_ID}"
        ;;
    esac
  )
}

function enable_and_start_services {
  local units=(
    ecms.service
    proton-cr.service
  )
  systemctl enable --now "${units[@]}"
}

install_binaries
install_rpm_packages
patch_for_distro
enable_and_start_services
