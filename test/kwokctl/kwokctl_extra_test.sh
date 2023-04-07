#!/usr/bin/env bash
# Copyright 2023 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

DIR="$(dirname "${BASH_SOURCE[0]}")"

DIR="$(realpath "${DIR}")"

RELEASES=()

EXTRASDIR="./extras"

function usage() {
  echo "Usage: $0 <kube-version...>"
  echo "  <kube-version> is the version of kubernetes to test against."
}

function args() {
  if [[ $# -eq 0 ]]; then
    usage
    exit 1
  fi
  while [[ $# -gt 0 ]]; do
    RELEASES+=("${1}")
    shift
  done
}

function show_info() {
    local name="${1}"
    echo kwokctl get clusters
    kwokctl get clusters
    echo
    echo kwokctl --name="${name}" kubectl get pod -o wide --all-namespaces
    kwokctl --name="${name}" kubectl get pod -o wide --all-namespaces
    echo
    echo kwokctl --name="${name}" logs etcd
    kwokctl --name="${name}" logs etcd
    echo
    echo kwokctl --name="${name}" logs kube-apiserver
    kwokctl --name="${name}" logs kube-apiserver
    echo
    echo kwokctl --name="${name}" logs kube-controller-manager
    kwokctl --name="${name}" logs kube-controller-manager
    echo
    echo kwokctl --name="${name}" logs kube-scheduler
    kwokctl --name="${name}" logs kube-scheduler
    echo
    echo kwokctl --name="${name}" logs kwok-controller
    kwokctl --name="${name}" logs kwok-controller
    echo
}

function test_create_cluster() {
  local release="${1}"
  local name="${2}"
  local targets
  local current_context
  local i

  KWOK_KUBE_VERSION="${release}" kwokctl --config "${DIR}/kwokctl-config-patches.yaml" create cluster --name "${name}" --timeout 10m --wait 10m --quiet-pull --prometheus-port 9090

  if [[ $? -ne 0 ]]; then
    echo "Error: Cluster ${name} creation failed"
    show_info "${name}"
    return 1
  fi
}

function test_delete_cluster() {
  local release="${1}"
  local name="${2}"
  kwokctl delete cluster --name "${name}"
}

function test_prometheus() {
  local targets
  for ((i = 0; i < 60; i++)); do
    targets="$(curl -s http://127.0.0.1:9090/api/v1/targets)"
    if [[ "$(echo "${targets}" | grep -o '"health":"up"' | wc -l)" -ge 6 ]]; then
      break
    fi
    sleep 1
  done

  if ! [[ "$(echo "${targets}" | grep -o '"health":"up"' | wc -l)" -ge 6 ]]; then
    echo "Error: metrics is not health"
    echo curl -s http://127.0.0.1:9090/api/v1/targets
    echo "${targets}"
    return 1
  fi
}

function prepare_mount_dirs() {
    mkdir "${EXTRASDIR}/apiserver"
    mkdir "${EXTRASDIR}/controller-manager"
    mkdir "${EXTRASDIR}/scheduler"
    mkdir "${EXTRASDIR}/controller"
    mkdir "${EXTRASDIR}/etcd"
    mkdir "${EXTRASDIR}/prometheus"
}

function main() {
  local failed=()
  local name

  mkdir -p "${EXTRASDIR}"
  prepare_mount_dirs
  for release in "${RELEASES[@]}"; do
    echo "------------------------------"
    echo "Testing extra on ${KWOK_RUNTIME} for ${release}"
    name="cluster-${KWOK_RUNTIME}-${release//./-}"
    test_create_cluster "${release}" "${name}" || failed+=("create_extra_cluster_${name}")
    test_prometheus || failed+=("prometheus_${name}")
    test_delete_cluster "${release}" "${name}" || failed+=("delete_extra_cluster_${name}")
  done
  echo "------------------------------"
  rm -rf "${EXTRASDIR}"

  if [[ "${#failed[@]}" -ne 0 ]]; then
    echo "------------------------------"
    echo "Error: Some tests failed"
    for test in "${failed[@]}"; do
      echo " - ${test}"
    done
    exit 1
  fi
}

args "$@"

main
