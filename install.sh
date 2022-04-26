#!/bin/sh

# Copyright (c) 2018 Bhojpur Consulting Private Limited, India. All rights reserved.
#
# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included in
# all copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
# OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
# THE SOFTWARE.

set -e
set -o noglob

github_version() {
    set +e
    curl -s https://api.github.com/repos/bhojpur/vpn/releases/latest | \
    grep tag_name | \
    awk '{ print $2 }' | \
    sed -e 's/\"//g' -e 's/,//g' || echo "v0.8.5"
    set -e
}

DOWNLOADER=${DOWNLOADER:-curl}
VERSION=${VERSION:-$(github_version)}

info()
{
    echo '[INFO] ' "$@"
}

warn()
{
    echo '[WARN] ' "$@" >&2
}

fatal()
{
    echo '[ERROR] ' "$@" >&2
    exit 1
}


detect_arch() {
    if [ -z "$ARCH" ]; then
        ARCH=$(uname -m)
    fi
    case $ARCH in
        i386)
            ARCH=i386
            ;;
        amd64|x86_64)
            ARCH=x86_64
            ;;
        arm64|aarch64)
            ARCH=arm64
            ;;
        arm*)
            ARCH=armv6
            ;;
        *)
            fatal "Unsupported architecture $ARCH"
    esac
}

detect_platform() {
    if [ -z "$OS" ]; then
        OS=$(uname -o)
    fi
    case $OS in
        *Linux)
            OS=Linux
            ;;
        *)
            fatal "Unsupported platform $OS"
    esac
}

verify_env() {

    detect_arch
    detect_platform

    if [ -x /sbin/openrc-run ]; then
        HAS_OPENRC=true
    fi
    if [ -x /bin/systemctl ] || type systemctl > /dev/null 2>&1; then
        HAS_SYSTEMD=true
    fi

    SUDO=sudo
    if [ $(id -u) -eq 0 ]; then
        SUDO=
    fi

    if [ -n "${INSTALL_BIN_DIR}" ]; then
        BIN_DIR=${INSTALL_BIN_DIR}
    else
        BIN_DIR=/usr/local/bin
        if ! $SUDO sh -c "touch ${BIN_DIR}/ro-test && rm -rf ${BIN_DIR}/ro-test"; then
            if [ -d /opt/bin ]; then
                BIN_DIR=/opt/bin
            fi
        fi
    fi

}

setup_service() {
    if [ -n "${INSTALL_SYSTEMD_DIR}" ]; then
        SYSTEMD_DIR="${INSTALL_SYSTEMD_DIR}"
    else
        SYSTEMD_DIR=/etc/systemd/system
    fi

    if [ "${HAS_SYSTEMD}" = true ]; then
        FILE_SERVICE=${SYSTEMD_DIR}/vpn@.service
        $SUDO tee $FILE_SERVICE >/dev/null << EOF
[Unit]
Description=Bhojpur VPN Daemon
After=network.target

[Service]
EnvironmentFile=/etc/systemd/system.conf.d/vpn-%i.env
LimitNOFILE=49152
ExecStartPre=-/bin/sh -c "sysctl -w net.core.rmem_max=2500000"
ExecStart=$BIN_DIR/vpnsvr
Restart=always

[Install]
WantedBy=multi-user.target
EOF
    elif [ "${HAS_OPENRC}" = true ]; then
        $SUDO tee /etc/init.d/vpn >/dev/null << EOF
#!/sbin/openrc-run
depend() {
    after network-online
}

supervisor=supervise-daemon
name=vpn
command="${BIN_DIR}/vpnsvr"
command_args="$(escape_dq "vpnsvr")
    >>${LOG_FILE} 2>&1"
output_log=${LOG_FILE}
error_log=${LOG_FILE}
pidfile="/var/run/vpnsvr.pid"
respawn_delay=5
respawn_max=0
set -o allexport
if [ -f /etc/environment ]; then source /etc/environment; fi
if [ -f /etc/bhojpur/vpn.env ]; then source /etc/bhojpur/vpn.env; fi
set +o allexport
EOF
    fi

}

download() {
    [ $# -eq 2 ] || fatal 'download needs exactly two arguments'

    case $DOWNLOADER in
        curl)
            curl -o $1 -sfL $2
            ;;
        wget)
            wget -qO $1 $2
            ;;
        *)
            fatal "Incorrect executable '$DOWNLOADER'"
            ;;
    esac

    # Abort if download command failed
    [ $? -eq 0 ] || fatal 'Download failed'
}

install() {
    info "Arch: $ARCH. OS: $OS Version: $VERSION (\$VERSION)"

    TMP_DIR=$(mktemp -d -t vpn-install.XXXXXXXXXX)

    download $TMP_DIR/out.tar.gz https://github.com/bhojpur/vpn/releases/download/$VERSION/vpnsvr-$VERSION-$OS-$ARCH.tar.gz

    # TODO verify w/ checksum
    tar xvf $TMP_DIR/out.tar.gz -C $TMP_DIR

    $SUDO cp -rf $TMP_DIR/vpnsvr $BIN_DIR/

    # TODO trap
    rm -rf $TMP_DIR

    # TODO setup env files for a network connection
}

verify_env
setup_service
install