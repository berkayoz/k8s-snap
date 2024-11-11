#!/bin/bash -xeu

DIR="$(realpath `dirname "${0}"`)"
cd "${DIR}/../../../.."

snapcraft clean k8s-binaries --use-lxd
snapcraft --use-lxd

if ! lxc profile show delve-proxy-profile
then
    lxc profile copy default delve-proxy-profile
fi
lxc profile edit delve-proxy-profile < "${DIR}/debug-proxy-profile.yaml"

if ! lxc profile show k8s-integration
then
    lxc profile copy default k8s-integration
fi
lxc profile edit k8s-integration < "${PWD}/tests/integration/lxd-profile.yaml"

lxc launch ubuntu:22.04 k8s-snap-debug --profile default --profile k8s-integration --profile delve-proxy-profile

trap "lxc delete k8s-snap-debug --force" EXIT

lxc file push $PWD/k8s_*.snap k8s-snap-debug/root/k8s.snap

lxc exec k8s-snap-debug -- bash -c "snap wait system seed.loaded"
lxc exec k8s-snap-debug -- bash -c "snap install go --classic"
lxc exec k8s-snap-debug -- bash -c "go install github.com/go-delve/delve/cmd/dlv@latest" 
lxc exec k8s-snap-debug -- bash -c "snap install /root/k8s.snap --classic --dangerous"
lxc exec k8s-snap-debug -- bash -c "/root/go/bin/dlv attach \$(pgrep k8sd) --continue --listen=:2345 --headless=true --log=true --log-output=debugger,debuglineerr,gdbwire,lldbout,rpc --accept-multiclient --api-version=2"
