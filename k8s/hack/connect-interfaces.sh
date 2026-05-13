#!/usr/bin/env bash

set -ue

if [ "$EUID" -ne 0 ]
then echo "Please run this script as root."
  exit 1
fi

declare -a INTERFACES=(
  cifs-mount
  cilium-module-load
  docker-privileged
  docker-unprivileged
  firewall-control
  fuse-support
  hardware-observe
  home
  home-read-all
  kernel-crypto-api
  kernel-module-observe
  kubernetes-support
  log-observe
  login-session-observe
  mount-observe
  network
  network-bind
  network-control
  network-observe
  opengl
  process-control
  system-observe
)

for if in "${INTERFACES[@]}"; do
  bash -c "snap connect 'k8s:${if}'"
done
