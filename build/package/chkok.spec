Name: chkok
Version: 0.3.1
Release: 1%{?dist}
Summary: checks attributes of files and sockets

License: MIT
URL: https://github.com/farzadghanei/chkok
Source0: %{name}-%{version}.tar.gz

# will use official golang tarballs instead, until go 1.22 rpm is in most repos
# BuildRequires: golang > 1.22, golang-gopkg-yaml-3-devel > 3.0.0

%description
"chkok" checks if attributes of files and sockets match the provided conditions to ensure
system state is as expected. It can be used to monitor system health and resource availability.
Provides different running modes, useful for reporting to local and remote monitoring.

# go toolchain stores go build id in a different ELF note than GNU toolchain
# so RPM can't find the build id from the binaries after build.
# https://github.com/rpm-software-management/rpm/issues/367
%global _missing_build_ids_terminate_build 0
%define debug_package %{nil}

%prep
%setup -c -q

%build
%make_build


%install
rm -rf $RPM_BUILD_ROOT
%make_install


%clean
rm -rf $RPM_BUILD_ROOT


%files
%license LICENSE
%doc README.rst
%{_bindir}/%{name}
%{_mandir}/man1/%{name}*
%{_docdir}/chkok/example-config.yaml


%changelog
* Sun Jun 16 2024 Farzad Ghanei 0.3.1-1
- Add support for Debian packaging
- Improve packaging

* Sat May 11 2024 Farzad Ghanei 0.3.0-1
- Add support for required headers for http runner
- Add maxHeaderBytes configuration for http runner
