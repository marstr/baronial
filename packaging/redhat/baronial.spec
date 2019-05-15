Name: baronial
Version: %{rpm_version}
Release: %{release}%{?dist}
Summary: Scriptable personal finance tooling
License: GPLv3
Source0: ./baronial-%{rpm_version}.tar.gz

%define debug_package %{nil}

BuildRequires: make go git

%prep
%setup -q -n baronial-%{raw_version}

%build
make bin/linux/baronial

%install
mkdir -p %{buildroot}/usr/bin
install -m 755 bin/linux/baronial %{buildroot}/usr/bin/baronial

%files
%license LICENSE
/usr/bin/baronial

%description
Command line budgeting application.
