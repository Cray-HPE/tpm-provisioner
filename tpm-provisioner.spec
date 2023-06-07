#
# MIT License
#
# (C) Copyright 2023 Hewlett Packard Enterprise Development LP
#
# Permission is hereby granted, free of charge, to any person obtaining a
# copy of this software and associated documentation files (the "Software"),
# to deal in the Software without restriction, including without limitation
# the rights to use, copy, modify, merge, publish, distribute, sublicense,
# and/or sell copies of the Software, and to permit persons to whom the
# Software is furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included
# in all copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
# THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
# OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
# ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
# OTHER DEALINGS IN THE SOFTWARE.
#
Name: tpm-provisioner-client
License: MIT
Summary: TPM Provisioner Client
Version: %(echo $VERSION | sed 's/^v//')
Release: %(echo $BUILD)
Source: %{name}-%{version}.tar.bz2
Group: Applications/System
Vendor: Hewlett Packard Enterprise Company

%ifarch %ix86
    %global GOARCH 386
%endif
%ifarch aarch64
    %global GOARCH arm64
%endif
%ifarch x86_64
    %global GOARCH amd64
%endif

%description
TPM Provisioner client

%prep
%setup -q

%build
make build-jenkins

%install
mkdir -p %{buildroot}/usr/local/bin
mkdir -p %{buildroot}/etc/tpm-provisioner
install -D -m 0644 conf/blobs.conf %{buildroot}/etc/tpm-provisioner/blobs.conf
install -D -m 0644 conf/client.conf %{buildroot}/etc/tpm-provisioner/client.conf
install -D -m 0755 bin/tpm-provisioner-client %{buildroot}/usr/local/bin/tpm-provisioner-client
install -D -m 0755 bin/tpm-getEK %{buildroot}/usr/local/bin/tpm-getEK
install -D -m 0755 bin/tpm-blob-clear %{buildroot}/usr/local/bin/tpm-blob-clear
install -D -m 0755 bin/tpm-blob-store %{buildroot}/usr/local/bin/tpm-blob-store
install -D -m 0755 bin/tpm-blob-retrieve %{buildroot}/usr/local/bin/tpm-blob-retrieve

%clean
rm -rf %{buildroot}

%files
%defattr(-,root,root)
%config(noreplace) /etc/tpm-provisioner/blobs.conf
%config(noreplace) /etc/tpm-provisioner/client.conf
/usr/local/bin/tpm-provisioner-client
/usr/local/bin/tpm-getEK
/usr/local/bin/tpm-blob-clear
/usr/local/bin/tpm-blob-store
/usr/local/bin/tpm-blob-retrieve
