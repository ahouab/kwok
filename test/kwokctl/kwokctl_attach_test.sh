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

LOGDIR="./logs"

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

function test_create_cluster() {
  local release="${1}"
  local name="${2}"

  KWOK_KUBE_VERSION="${release}" kwokctl -v=-4 create cluster --name "${name}" --timeout 10m --wait 10m --quiet-pull --config="${DIR}/attach.yaml"
  if [[ $? -ne 0 ]]; then
    echo "Error: Cluster ${name} creation failed"
    exit 1
  fi
}

function test_attach() {
  local name="${1}"
  local namespace="${2}"
  local target="${3}"
  local targetLog="${LOGDIR}/kwok.log"

  echo '2016-10-06T00:17:09.669794202Z stdout F log content 1' > "${targetLog}"
  echo '2016-10-06T00:18:09.669794202Z stdout F log content 2' >> "${targetLog}"
  echo '2016-10-06T00:19:09.669794202Z stdout F log content 3' >> "${targetLog}"

  local attachLog="${LOGDIR}/attach.out"
  kwokctl --name "${name}" kubectl -n "${namespace}" attach "${target}" > "${attachLog}" &
  pid=$!

  # allow some time for attach to parse logs
  sleep 1

  echo '2016-10-06T00:20:09.669794202Z stdout F log content 4' >> "${targetLog}"
  echo '2016-10-06T00:20:10.669794202Z stdout F log content 5' >> "${targetLog}"

  local want=$(tail -n 2 "${targetLog}" | cut -d " " -f 4-)

  local result
  for ((i = 0; i < 60; i++)); do
    result=$(cat "${attachLog}")
    if [[ "${result}" =~ "${want}" ]]; then
        break
    fi
    sleep 1
  done

  kill -INT "${pid}"

  result=$(cat "${attachLog}")
  if [[ ! "${result}" =~ "${want}" ]]; then
      echo "Error: attach result does not match"
      echo " want: ${want}"
      echo " got: ${result}"
      return 1
  fi
}

function test_apply_node_and_pod() {
  kwokctl --name "${name}" kubectl apply -f "${DIR}/fake-node.yaml"
  if [[ $? -ne 0 ]]; then
    echo "Error: fake-node apply failed"
    return 1
  fi
  kwokctl --name "${name}" kubectl apply -f "${DIR}/fake-pod-in-other-ns.yaml"
  if [[ $? -ne 0 ]]; then
    echo "Error: fake-pod apply failed"
    return 1
  fi
  kwokctl --name "${name}" kubectl apply -f "${DIR}/fake-deployment.yaml"
  if [[ $? -ne 0 ]]; then
    echo "Error: fake-deployment apply failed"
    return 1
  fi
  kwokctl --name "${name}" kubectl wait pod -A --all --for=condition=Ready --timeout=60s
  if [[ $? -ne 0 ]]; then
    echo "Error: fake-pod wait failed"
    echo kwokctl --name "${name}" kubectl get pod -A --all
    kwokctl --name "${name}" kubectl get pod -A --all
    return 1
  fi
}

function test_delete_cluster() {
  local release="${1}"
  local name="${2}"
  kwokctl delete cluster --name "${name}"
}

function main() {
  local failed=()
  mkdir -p "${LOGDIR}"

  for release in "${RELEASES[@]}"; do
    echo "------------------------------"
    echo "Testing attaches on ${KWOK_RUNTIME} for ${release}"
    name="attach-cluster-${KWOK_RUNTIME}-${release//./-}"
    test_create_cluster "${release}" "${name}" || failed+=("create_cluster_${name}")
    test_apply_node_and_pod || failed+=("apply_node_and_pod")
    test_attach "${name}" other pod/fake-pod || failed+=("${name}_target_attaches")
    test_attach "${name}" default deploy/fake-pod || failed+=("${name}_cluster_default_attaches")
    test_delete_cluster "${release}" "${name}" || failed+=("delete_cluster_${name}")
  done
  rm -rf "${LOGDIR}"

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
