#!/bin/bash
SCRIPT_PATH=`dirname $(realpath $0)`

eceph_script="${SCRIPT_PATH}/scripts/install-eceph-deps.sh"
if [ -f "$eceph_script" ]; then
  source "$eceph_script"
fi

HOSTS=""
PASSWD=""
User="root"
Port="22"
WITH_ECEPH_REMOTE=""
WITH_ECEPH=false
WITH_NVIDIA_DRIVER=false
CONTAINER_RUNTIME="docker"
para_support_list=(`echo "--help --hosts --ssh-password --ssh-user --ssh-port --with-eceph-remote --with-eceph --with-nvidia-driver --container-runtime" | tr '=' ' ' `)

Usage()
{
    echo "Usage:    install_deps [arg...]"
    echo "  install_deps.sh [ --help ]"
    echo
    echo "Multi-node installation proton-cli dependency package."
    echo
    echo "Options:"
    echo
    echo "    -h, --help                     Output help information"
    echo "    --hosts  string                Node list, multiple nodes are divided by , None means local installation on this node only"
    echo "    --ssh-password  string         Node ssh password, multi-node password is the same"
    echo "    --ssh-user  string             Node ssh user name, multi-node user name is the same, default root"
    echo "    --ssh-port  string             Node ssh port, multi-node port is the same, default 22"
    echo "    --with-eceph-remote string     Node list to install ECeph, multiple nodes are divided by ,  None means do not install ECeph"
    echo "    --with-eceph boolean           Whether to install ECeph on this node,  false means do not install ECeph, should not be used with with-eceph-remote"
    echo "    --with-nvidia-driver boolean   Whether to install Nvidia on this node,  false means do not install nvidia driver"
    echo "    --container-runtime {docker|containerd}"
    echo "                                   Specify the container runtime for kubernetes (default \"docker\")"

    echo
    exit 1
}
#Analytic Functions
function ParseCommand()
{

  for para in "$@"; do
    if [[ $para =~ "--"  ]]; then
      is_existe=false
      for l in "${para_support_list[@]}"; do
        if [[ "$l" == "$para" ]];then
          is_existe=true
        fi
      done
      if [[ "${is_existe}" == "false" ]];then
        echo "$para is not supported, the allowed values are [${para_support_list[@]}]" && exit 1
      fi
    fi
  done
	ARGS=`getopt -o h --long help,hosts:,ssh-password:,ssh-user:,ssh-port:,with-eceph-remote:,with-eceph:,with-nvidia-driver:,container-runtime: -- "$@"`
	[ $? -ne 0 ] && Usage
  if [[ ! `echo ${ARGS}  | awk '{print substr($0,length,1)}' ` == '-'  ]];then
    echo "wrong parameter, please check ${ARGS}  " && exit 1
  fi
	set -- "${ARGS}"
	eval set -- "${ARGS}"
	while true
	do
		case "$1" in
		-h|--help)
				Usage
				;;
		--hosts)
				HOSTS=$2
				shift
				;;
		--ssh-password)
				PASSWD=$2
				shift
				;;
		--ssh-user)
				User=$2
				shift
				;;
		--ssh-port)
				Port=$2
				shift
				;;
    --with-eceph-remote)
        WITH_ECEPH_REMOTE=$2
        shift
        ;;
    --with-eceph)
        WITH_ECEPH=$2
        shift
        ;;
    --with-nvidia-driver)
        WITH_NVIDIA_DRIVER=$2
        shift
        ;;
    --container-runtime)
        CONTAINER_RUNTIME=$2
        shift
        ;;
		--)
				shift
				break
				;;
		esac
	shift
	done
  if [ "${WITH_ECEPH}" != "false" ] && [ "${WITH_ECEPH_REMOTE}" != "" ];then
    echo "with-eceph and with-eceph-remote should not be used at the same time" && exit 1
  fi
  if [ "${HOSTS}" == "" ] && [ "${WITH_ECEPH_REMOTE}" != "" ];then
    echo "hosts and with-eceph-remote must be used together" && exit 1
  fi
  if [ "${HOSTS}" != "" ] && [ "${WITH_ECEPH}" != "false" ];then
    echo "hosts and with-eceph cannot be used together, use with-eceph-remote for remote installation" && exit 1
  fi
  if [ -d ./service-package-eceph ] && [ -d ./repos/eceph ]; then
    echo "ECeph repo packages found"
  else
    if [ "${WITH_ECEPH_REMOTE}" != "" ] || [ "${WITH_ECEPH}" != "false" ]; then
      echo "please download and unpack ECeph repo package first before installing ECeph" && exit 1
    fi
  fi
  if [ "${WITH_NVIDIA_DRIVER}" != "false" ] && [[ "$(arch)" = "aarch64" ]]; then
    echo "Nvidia driver is not supported on aarch64" && exit 1
  fi
  echo $HOSTS $User $PASSWD $Port
}

function fix_docker_config_login_nil()
{
  docker_config_file=/root/.docker/config.json
  if [ -f $docker_config_file ]; then
    echo "fix docker config login nil error"
    sed -i 's/Og==//g' $docker_config_file
  fi
}

function disable_networkmanager()
{
  nm_active=$(systemctl is-active NetworkManager)
  network_active=$(systemctl is-active network)
  if [ "$nm_active" == "active" ] && [ "$network_active" == "active" ];then
    echo "disable NetworkManager"
    systemctl stop NetworkManager
    systemctl disable NetworkManager
  fi
}

function networkmanager_unmanage_docker0()
{
  nm_active=$(systemctl is-active NetworkManager)
  if [ "$nm_active" == "active" ];then
    if [ -f /etc/NetworkManager/conf.d/98-dns-none.conf ];then
    # check if exist dns=none in 98-dns-none.conf
      if grep -q dns /etc/NetworkManager/conf.d/98-dns-none.conf; then
        echo "dns=none already unmanaged in 98-dns-none.conff"
      else
        echo "WARN:YOU SHOULD ADD dns=none to /etc/NetworkManager/conf.d/98-dns-none.conf MANUAL!!!"
      fi
    else
      # else create file and add [main]\ndns=none
      echo "[main]" > /etc/NetworkManager/conf.d/98-dns-none.conf
      echo "dns=none" >> /etc/NetworkManager/conf.d/98-dns-none.conf
      echo "add dns=none unmanaged to /etc/NetworkManager/conf.d/98-dns-none.conf"
    fi

    if [ -f /etc/NetworkManager/conf.d/99-unmanaged-devices.conf ];then
    # check if exist docker0 in unmanaged-devices.conf
      if grep -q docker0 /etc/NetworkManager/conf.d/99-unmanaged-devices.conf; then
        echo "docker0 already unmanaged in 99-unmanaged-devices.conf"
      else
        echo "WARN:YOU SHOULD ADD docker0 to unmanaged-devices.conf MANUAL!!!"
      fi
    else
      # else create file and add [keyfile]\nunmanaged-devices=interface-name:docker0
      echo "[keyfile]" > /etc/NetworkManager/conf.d/99-unmanaged-devices.conf
      echo "unmanaged-devices=interface-name:docker0" >> /etc/NetworkManager/conf.d/99-unmanaged-devices.conf
      systemctl reload NetworkManager
      echo "add docker0 unmanaged to 99-unmanaged-devices.conf"
    fi
  fi
}

# install libxcrypt-compat on almalinux 9
function install_libxcrypt_compat() {
  if [ -f /etc/almalinux-release ];then
  # check /etc/almalinux-release contains "AlmaLinux release 9"
    if grep -q "AlmaLinux release 9" /etc/almalinux-release;then
      if ! rpm -q libxcrypt-compat >/dev/null 2>&1;then
        echo "install libxcrypt-compat on almalinux 9"
        rpm -Uvh ./scripts/libxcrypt-compat*.rpm
      fi
    fi
  fi
}

Tranfer_Package_And_Install_Rpm() {
	echo "===Tranfer_Package_And_Install_proton==="
  SHOULD_INSTALL_ECEPH_HERE=false;
  ip=$1
  ECEPH_IPS=(`echo ${WITH_ECEPH_REMOTE} | tr ',' ' '`)
  for ecephip in "${ECEPH_IPS[@]}"; do
    if [[ "${ecephip}" == "${ip}" ]]; then
      SHOULD_INSTALL_ECEPH_HERE=true;
      echo "should install ECeph on node ${ip}";
    fi;
  done;

  if [[ ${PASSWD} == "" ]]; then
    ssh -p ${Port}    -o StrictHostKeyChecking=no  ${User}@${ip} "mkdir -p /tmp/proton-package 1>/dev/null && rm -rf /tmp/proton-package/* " || exit 1
    scp -P ${Port} -r -o StrictHostKeyChecking=no  $SCRIPT_PATH/* ${User}@[${ip}]:/tmp/proton-package && echo ${ip}:Transfer_Package SUCCESS || exit 1
    ssh -p ${Port}    -o StrictHostKeyChecking=no  ${User}@${ip} " cd /tmp/proton-package && bash install_deps.sh --with-eceph ${SHOULD_INSTALL_ECEPH_HERE} --container-runtime ${CONTAINER_RUNTIME} && rm -rf /tmp/proton-package" || exit 1
  else
    $SCRIPT_PATH/scripts/sshpass -p ${PASSWD} ssh -p ${Port} -o StrictHostKeyChecking=no  ${User}@${ip} "mkdir -p /tmp/proton-package 1>/dev/null && rm -rf /tmp/proton-package/* " || exit 1
    $SCRIPT_PATH/scripts/sshpass -p ${PASSWD} scp -P ${Port} -r -o StrictHostKeyChecking=no  $SCRIPT_PATH/* ${User}@[${ip}]:/tmp/proton-package && echo ${ip}:Transfer_Package SUCCESS || exit 1
    $SCRIPT_PATH/scripts/sshpass -p ${PASSWD} ssh -p ${Port} -o StrictHostKeyChecking=no  ${User}@${ip} " cd /tmp/proton-package && bash install_deps.sh --with-eceph ${SHOULD_INSTALL_ECEPH_HERE} --container-runtime ${CONTAINER_RUNTIME} && rm -rf /tmp/proton-package" || exit 1
  fi
    # sshpass -p fake_pass scp -P 22 test.json root@[${ip}]:/tmp
}

function check_resolv()
{
  if [[ ! -f /etc/resolv.conf ]]; then
    echo "nameserver 8.8.8.8" >> /etc/resolv.conf
  fi
  if [[ `cat /etc/resolv.conf | grep openstacklocal | wc -l` -gt 0 ]]; then
    sed -i "/openstacklocal/d" /etc/resolv.conf
  fi
  if [[ `cat /etc/resolv.conf | grep nameserver | wc -l` -eq 0 ]]; then
     echo "nameserver 8.8.8.8" >> /etc/resolv.conf
  fi

}
function check_hosts()
{
  if [[ ! -f /etc/hosts ]]; then
    echo "127.0.0.1   localhost localhost.localdomain localhost4 localhost4.localdomain4" >> /etc/hosts
    echo "::1         localhost localhost.localdomain localhost6 localhost6.localdomain6" >> /etc/hosts
  fi
  if [[ `cat /etc/hosts | grep "127.0.0.1[[:space:]]*localhost" | wc -l` -eq 0 ]]; then
     echo "127.0.0.1   localhost localhost.localdomain localhost4 localhost4.localdomain4" >> /etc/hosts
  fi
  if [[ `cat /etc/hosts | grep "::1[[:space:]]*localhost" | wc -l` -eq 0 ]]; then
     echo "::1         localhost localhost.localdomain localhost6 localhost6.localdomain6" >> /etc/hosts
  fi
}

# Add keepalived detection script
function Add_keepalived_script()
{
  if ! [[ -d /etc/keepalived ]]; then
    mkdir -p /etc/keepalived
  fi
  if [[ -e /etc/keepalived/nginx_check.sh ]];then
      nginx_port=`sed -n 's/.*nginx_port=\([0-9]*\).*/\1/p' /etc/keepalived/nginx_check.sh`
      if ! [[ -n $nginx_port ]];then
          \cp -rf $SCRIPT_PATH/scripts/nginx_check.sh /etc/keepalived
      fi
  else
      \cp -rf $SCRIPT_PATH/scripts/nginx_check.sh /etc/keepalived
  fi

}
#install yaml tool yq
Install_yq() {
  \cp -rf "${SCRIPT_PATH}/scripts/yq" /usr/bin/yq
  \cp -rf "${SCRIPT_PATH}/scripts/yq" /usr/local/bin/yq
  chmod +x /usr/bin/yq
  chmod +x /usr/local/bin/yq
	echo "Install yq Success"
}
# upgrade nssswitch.config
function upgrde_nss_switch()
{
  if [[ -f "/etc/nsswitch.conf" ]]; then
    sed -i "/^passwd:.*/c\passwd:      files systemd sss" /etc/nsswitch.conf
    sed -i "/^group:.*/c\group:       files systemd sss" /etc/nsswitch.conf
  fi
}

function handle_firewalld()
{
  if systemctl is-active --quiet firewalld; then
    echo "firewalld is active, opening port 9202/tcp"
    firewall-cmd --permanent --add-port=9202/tcp
    firewall-cmd --reload
  else
    echo "WARN: firewalld is not active. Please ensure port 9202/tcp is accessible if you enable it later."
  fi
}
# sysctl
function sysctl_default()
{
echo """net.ipv4.tcp_sack = 1
net.ipv4.tcp_synack_retries = 3
vm.max_map_count = 262144
kernel.shmall = 4294967296
kernel.msgmnb = 65536
net.ipv4.icmp_echo_ignore_broadcasts = 1
net.core.somaxconn = 10240
fs.aio-max-nr = 262144
net.ipv4.ip_local_reserved_ports = 10001,10002,10250,14322,19122,7480,14322,19122,9001,9065,9703,9123,9998,8001,9028,9027,10031,9996,10028,30002,10025,9300,9302,16443,18080,18008
net.ipv4.conf.default.rp_filter = 0
net.ipv4.tcp_syn_retries = 3
net.ipv4.tcp_fin_timeout = 60
net.ipv4.ip_local_port_range = 15000 65000
net.netfilter.nf_conntrack_tcp_timeout_established = 3600
net.netfilter.nf_conntrack_max = 524288
kernel.shmmax = 68719476736
net.ipv4.conf.default.accept_source_route = 0
kernel.core_uses_pid = 1
net.ipv6.route.max_size = 2147483647
kernel.msgmax = 65536
net.ipv4.tcp_syncookies = 1
net.ipv4.tcp_max_syn_backlog = 32768
vm.swappiness = 0
vm.dirty_ratio = 30
net.core.netdev_max_backlog = 3000
kernel.sysrq = 0
net.ipv4.tcp_tw_reuse = 1""" > /etc/sysctl.d/proton.conf
modprobe nf_conntrack
modprobe ip_conntrack
sysctl -p /etc/sysctl.d/proton.conf
}

# install version file
function save_version_file()
{
  echo "Now version: $(cat ${SCRIPT_PATH}/proton-version.txt)"
  if [[ ! -d /usr/local/share ]]; then
    mkdir -p /usr/local/share
  fi
  \cp -rf "${SCRIPT_PATH}/proton-version.txt" /usr/local/share/proton-version.txt
}

#Install dependencies
function install_deps()
{
  if [[ "${CONTAINER_RUNTIME}" == "docker" ]]; then
    fix_docker_config_login_nil || exit 1
  fi
  disable_networkmanager || exit 1
  networkmanager_unmanage_docker0 || exit 1
  check_resolv || exit 1
  check_hosts || exit 1
  Install_yq  || exit 1
  upgrde_nss_switch || exit 1
  handle_firewalld || exit 1
  sysctl_default || exit 1
  proton_deps_list=(
    proton-cs
    "proton-slb-1.2.8-127.el7"
    proton-cr
    ecms
    rsync
    chrony
    proton-healthcheck
    anyshare_tools
    deploy_tools
  )
  case "${CONTAINER_RUNTIME}" in
    containerd)
      proton_deps_list+=(containerd nerdctl buildkit)
      ;;
    docker)
      proton_deps_list+=(docker-ce docker-ce-cli)
      ;;
    *)
      echo "unsupported container runtime, supported values: containerd, docker" > /dev/stderr
      ;;
  esac
  install_libxcrypt_compat || exit 1
  echo "install proton deps begin."

  # reinstall wrong version ecms, https://devops.aishu.cn/AISHUDevOps/ICT/_workitems/edit/586593/
  if rpm -q "ecms" >/dev/null 2>&1;then
    version=$(rpm -q --queryformat='%{VERSION}' ecms)
    if [[ $version =~ ^[0-9]+$ ]];then
      rpm -e --nodeps ecms
    fi
  fi

  # RPM Repository 配置文件
  sed "s|BASEURL|file://${SCRIPT_PATH}/repos|" "${SCRIPT_PATH}/repos/proton-package.repo.tmpl" > "${SCRIPT_PATH}/repos/proton-package.repo"

  readonly DNF="dnf"
  readonly YUM="yum"

  # Choose yum or dnf as package management tool
  # Outputs:
  #   Write `yum` or `dnf` to stdout.
  function choose_package_manager_yum_or_dnf {
    readonly local YUM_MAXIMUM_VERSION="3.4.3"
    yum_version=$(yum --version | head -1)
    if [[ "$(printf "%s\n%s\n" "${YUM_MAXIMUM_VERSION}" "${yum_version}" | tail -1)" == "${YUM_MAXIMUM_VERSION}" ]]; then
      echo "${YUM}"
    else
      echo "${DNF}"
    fi
  }

  # TODO: Remove these after installing on environments where proton-cs lower than 1.9.1 is not supported.
  # Update  proton-cs not executing package scriptlet(s)
  function update_proton_cs_not_exec_rpm_scriptlets {
    local proton_cs_maximum_version="1.9.1"
    local regex_proton_cs_and_proton_cs_images='/tmp/repos/Packages/proton-cs\(\|-images\)-[1-9][0-9]*\.[0-9]+\.[0-9]+-[1-9][0-9]+\.el7\.\(x86_64\|aarch64\)\.rpm'
    if rpm -q proton-cs >/dev/null; then
      proton_cs_version="$(rpm -q proton-cs --queryformat %{VERSION})"
      if [[ "$(printf "${proton_cs_version}\n${proton_cs_maximum_version}\n" | sort --version-sort | head -1)" != "${proton_cs_maximum_version}" ]]; then
        local -a proton_cs_and_proton_images
        for pkg in $(find /tmp/repos/Packages -type f -regex "${regex_proton_cs_and_proton_cs_images}"); do
          proton_cs_and_proton_images+=( "${pkg}" )
        done
        rpm -U --noscripts "${proton_cs_and_proton_images[@]}"
      fi
    fi
  }

  if grep -ic "suse" /etc/os-release
  then
    update_proton_cs_not_exec_rpm_scriptlets
    zypper ar -fcG "${SCRIPT_PATH}/repos" proton-package
    # system MUST INSTALLED chrony and firewalld
    zypper install --auto-agree-with-licenses -y --repo=proton-package ${proton_deps_list[*]} || :
    zypper rr proton-package
    # remove proton-slb and proton-cr bin/libreadline.so.6, use system default /lib64/libreadline.so.6.3
    rm -f /usr/local/proton-cr/bin/libreadline.so.6
    rm -f /usr/local/slb/bin/libreadline.so.6
    rm -f /usr/local/ecms/bin/libreadline.so.6
  else
    # uninstall anyshare-deps and proton-deps
    pkgs=(
      anyshare-deps
      anyshare-tools
      deploytools
    )
    # kylin v10 sp2 runc conflict with proton runc, remove first
    # rhel 8.6 runc conflict with proton runc
    kylinv10_pkgs=(
      docker-runc
      docker-engine
    )
    opts=(
      -y
    )

    pkg_manager=$(choose_package_manager_yum_or_dnf)
    if [[ "${pkg_manager}" == "${DNF}" ]]; then
      opts+=( --noautoremove )
    fi

    if [[ "${pkg_manager}" == "${YUM}" ]]; then
      yum clean all
    fi

    "${pkg_manager}" "${opts[@]}" remove "${pkgs[@]}"
    "${pkg_manager}" "${opts[@]}" remove "${kylinv10_pkgs[@]}"
    # remove proton-cs after anyshare-deps because anyshare-deps requires proton-cs
    # 当proton-cs-1.8.1时，卸载proton-cs失败，因为cr-tools-1.25依赖它，这里先注释，现在看上去不需要卸载了
    # update_proton_cs_not_exec_rpm_scriptlets
    "${pkg_manager}" --config="${SCRIPT_PATH}/repos/proton-package.repo" --disablerepo="*" --assumeyes repository-packages proton-package install "${proton_deps_list[@]}"
    # upgrade ecms
    "${pkg_manager}" --config="${SCRIPT_PATH}/repos/proton-package.repo" --disablerepo="*" --assumeyes repository-packages proton-package  upgrade ecms
  fi

  # value of net.ipv4.ip_forward in OpenEuler and Kylin V10 is default 0
sed -i '/net.ipv4.ip_forward/s|net.ipv4.ip_forward[[:space:]]*=.*|net.ipv4.ip_forward=1|g' /etc/sysctl.d/*
sed -i '/net.ipv4.ip_forward/s|net.ipv4.ip_forward[[:space:]]*=.*|net.ipv4.ip_forward=1|g' /etc/sysctl.conf
  echo 1 > /proc/sys/net/ipv4/ip_forward

  systemctl enable ecms
  systemctl start ecms
  systemctl enable proton_slb_manager
  systemctl start proton_slb_manager

  rm -rf /usr/share/proton /usr/share/ab_tools
  \cp -rf scripts/ab_tools  /usr/share/ab_tools
  \cp -rf scripts/proton-cli /usr/bin
  \cp -rf scripts/skopeo /usr/bin/
  \cp -rf scripts/nebula-console /usr/bin/
  \cp -rf scripts/nic_bind_core.py /usr/bin/
  # 安装 slb-nginx 所依赖的 libpcre。slb-nginx 之后应改为静态编译。
  install -p -t /usr/local/slb-nginx/third_lib lib/libpcre.so.1
  Add_keepalived_script
  # Create scheduled backup tasks
	if [[ `crontab -l |grep "proton-cli" | wc -l` -eq 0 ]];then
    echo "Create scheduled backup tasks"
		proton-cli backup schedule create --schedule="0 2 * * * proton-cli backup create --resources all"
	fi

  if "${WITH_ECEPH}"; then
    eceph::install_deps "${SCRIPT_PATH}"
  fi

  if "$WITH_NVIDIA_DRIVER"; then
    if [ -d "/tmp/nvidia-driver" ]
    then
        rm -rf /tmp/nvidia-driver/*
    else
        mkdir -p /tmp/nvidia-driver
    fi
    tar xf nvidia-driver.tar.gz -C /tmp/nvidia-driver
    /tmp/nvidia-driver/nvidia-driver-*/install.sh
  fi

  helm3 -n resource uninstall installer-service > /dev/null 2>&1 || :
  save_version_file || exit 1

  # 读取操作系统信息
  if [[ -f /etc/os-release ]]; then
    source /etc/os-release
  else
    ID="unknown"
    VERSION_ID="unknown"
  fi
  # 禁用 SUSE Linux Enterprise Server 12 SP5 的 apparmor.service，否则无法启动 docker 容器。
  if [[ "${ID}" == "sles" ]] && [[ "${VERSION_ID}" == "12.5" ]]; then
    systemctl --now disable apparmor.service
  fi

  echo "install proton deps end."

}




Main() {
  ParseCommand $@
  set -ex
  # The node parameters are not set, execute the local installation proton-cli dependency
  if [[ ${HOSTS} == "" ]]; then
    install_deps
  else
    IPS=(`echo ${HOSTS} | tr ',' ' '`)
    for ip in "${IPS[@]}"; do
      Tranfer_Package_And_Install_Rpm $ip
    done
  fi
}


Main $@
