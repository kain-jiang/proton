%if %{getenv:BUILD_NUMBER}0
%define     build_number    %{getenv:BUILD_NUMBER}
%else
%define     build_number    1
%endif
%define _logdir /var/log/ecms

%define _arch %(echo %{dist} | cut -d '.' -f4)
%define _dist .el7%{_arch}

Name:    ecms
Version: ECMS_VERSION
Release: %{build_number}%{_dist}
Summary: ECMS Service
Group:   Applications/System
License: GPLv2
URL:     https://www.aishu.cn
Source:  package.tar.gz

Requires:   chrony
Requires:   firewalld
Requires:   iproute
Requires:   net-tools

%description
ECMS service.

%prep
%setup -q -n package

%build

%install
mkdir -p $RPM_BUILD_ROOT%{_sysconfdir}/ecms/
mkdir -p $RPM_BUILD_ROOT%{_prefix}/local/ecms
mkdir -p $RPM_BUILD_ROOT%{_unitdir}
mkdir -p $RPM_BUILD_ROOT%{_logdir}
cp -rpf %{_builddir}/package/bin $RPM_BUILD_ROOT%{_prefix}/local/ecms/
cp -rpf %{_builddir}/package/config/ecms.conf $RPM_BUILD_ROOT%{_sysconfdir}/ecms/
cp -rpf %{_builddir}/package/boost/ecms.service $RPM_BUILD_ROOT%{_unitdir}/


%post
%systemd_post ecms.service

%preun
%systemd_preun ecms.service

%postun
%systemd_postun_with_restart ecms.service


%files
%defattr (0755,root,root,0755)
%dir %{_sysconfdir}/ecms/
%config(noreplace) %{_sysconfdir}/ecms/*
%dir %{_logdir}
%dir %{_prefix}/local/ecms
%{_prefix}/local/ecms/*
%attr(644, root, root) %{_unitdir}/ecms.service


%changelog
