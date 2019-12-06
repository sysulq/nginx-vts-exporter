#! /bin/bash

self=`readlink -f -- "$0"`
rpm_build_root=`dirname -- "$self"`

cd "$rpm_build_root/.."

project_source_dir=`pwd -P`

VERSION=`cat $project_source_dir/VERSION`
TEMPLATE="$rpm_build_root/nginx_vts_exporter.spec.tmpl"
SRCRPM_OUT_DIR="$rpm_build_root/SRPMS"
RPM_OUT_DIR="$rpm_build_root/RPMS"


check_build_requires() {
    while [ $# -gt 0 ]; do
        if ! command -v "$1" 2>&1 >/dev/null; then
            echo -e "missing binary dependency: $1"
            return 1
        fi
        shift 1
    done
    return 0
}

generate_rpm_spec() {
    export nginx_vts_exporter_version=$VERSION
    envsubst <  "$TEMPLATE"
}

prepare_manifests() {
    mkdir -p "$SRCRPM_OUT_DIR"
    mkdir -p "$rpm_build_root/SPECS"
    mkdir -p "$rpm_build_root/SOURCES"

    rm -f "$rpm_build_root/SOURCES/nginx_vts_exporter-${VERSION}.tar.gz"
    tar -zcvf "$rpm_build_root/SOURCES/nginx_vts_exporter-${VERSION}.tar.gz" \
                --transform "s/^\./nginx-vts-exporter-${VERSION}/" \
                --exclude="SOURCES" \
                --exclude="SPECS" \
                --exclude=".git" \
                .
    generate_rpm_spec > "$rpm_build_root/SPECS/nginx_vts_exporter.spec"
    cp -f "$project_source_dir/systemd/nginx_vts_exporter.default" "$rpm_build_root/SOURCES/nginx_vts_exporter.default"
    cp -f "$project_source_dir/systemd/nginx_vts_exporter.service" "$rpm_build_root/SOURCES/nginx_vts_exporter.service"
}

build_srpm() {
    prepare_manifests || return 1

    rpmbuild -bs "$rpm_build_root/SPECS/nginx_vts_exporter.spec" \
                --define "%_srcrpmdir $SRCRPM_OUT_DIR" \
                --define "%_topdir $rpm_build_root"
}

build_rpm() {
    prepare_manifests || return 1
    yum-builddep "$rpm_build_root/SPECS/nginx_vts_exporter.spec" || return 1
    rpmbuild -bb "$rpm_build_root/SPECS/nginx_vts_exporter.spec" \
                --define "%_rpmdir $RPM_OUT_DIR" \
                --define "%_topdir $rpm_build_root"
}

build_package() {
    check_build_requires rpmbuild yum-builddep envsubst tar || return 1

    local pkg_type="$1"
    case ${pkg_type:=srpm} in
        srpm)
            build_srpm
            ;;
        rpm)
            build_rpm
            ;;
        *)
            echo -e "unknown package type: $pkg_type"
            return 1
            ;;
    esac
}

set -e
build_package $*

