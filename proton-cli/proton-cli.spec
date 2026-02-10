%if %{getenv:BUILD_NUMBER}0
%define     build_number    %{getenv:BUILD_NUMBER}
%else
%define     build_number    1
%endif

%define _arch %(echo %{dist} | cut -d '.' -f4)
%define _dist .el7%{_arch}

Name:    proton-cli
Version: PROTON_CLI_VERSION
Release: %{build_number}%{_dist}
Summary: Proton CLI
Group:   Applications/System
License: GPLv2
URL:     https://www.aishu.cn
Source:  proton-cli

%description
Proton kubernetes cluster and services deploy command line interface.


%build

%install
mkdir -p %{buildroot}/usr/bin/
install -m 755 /root/rpmbuild/SOURCES/proton-cli %{buildroot}/usr/bin/proton-cli


%files
/usr/bin/proton-cli


%changelog
* Thu Jun 9 2022 Proton Group
- release PROTON_CLI_VERSION
