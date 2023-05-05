/*
Copyright 2022 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package components

import (
	"fmt"
	"strconv"

	"sigs.k8s.io/kwok/pkg/apis/internalversion"
	"sigs.k8s.io/kwok/pkg/log"
	"sigs.k8s.io/kwok/pkg/utils/format"
	"sigs.k8s.io/kwok/pkg/utils/version"
)

// BuildKubeApiserverComponentConfig is the configuration for building a kube-apiserver component.
type BuildKubeApiserverComponentConfig struct {
	Binary            string
	Image             string
	Version           version.Version
	Workdir           string
	Address           string
	Port              uint32
	EtcdAddress       string
	EtcdPort          uint32
	KubeRuntimeConfig string
	KubeFeatureGates  string
	SecurePort        bool
	KubeAuthorization bool
	KubeAdmission     bool
	AuditPolicyPath   string
	AuditLogPath      string
	CaCertPath        string
	AdminCertPath     string
	AdminKeyPath      string
	Verbosity         int
	ExtraArgs         []internalversion.ExtraArgs
	ExtraVolumes      []internalversion.Volume
}

// BuildKubeApiserverComponent builds a kube-apiserver component.
func BuildKubeApiserverComponent(conf BuildKubeApiserverComponentConfig) (component internalversion.Component, err error) {
	if conf.EtcdPort == 0 {
		conf.EtcdPort = 2379
	}

	if conf.Address == "" {
		conf.Address = publicAddress
	}

	if conf.EtcdAddress == "" {
		conf.EtcdAddress = localAddress
	}

	kubeApiserverArgs := []string{
		"--etcd-prefix=/registry",
		"--allow-privileged=true",
	}

	if conf.KubeAdmission {
		if conf.Version.LT(version.NewVersion(1, 21, 0)) && !conf.KubeAuthorization {
			return component, fmt.Errorf("the kube-apiserver version is less than 1.21.0, and the --kube-authorization is not enabled, so the --kube-admission cannot be enabled")
		}
	} else {
		// TODO: use enable-admission-plugins and disable-admission-plugins instead of admission-control
		kubeApiserverArgs = append(kubeApiserverArgs,
			"--admission-control=",
		)
	}

	kubeApiserverArgs = append(kubeApiserverArgs, extraArgsToStrings(conf.ExtraArgs)...)
	if conf.KubeRuntimeConfig != "" {
		kubeApiserverArgs = append(kubeApiserverArgs,
			"--runtime-config="+conf.KubeRuntimeConfig,
		)
	}
	if conf.KubeFeatureGates != "" {
		kubeApiserverArgs = append(kubeApiserverArgs,
			"--feature-gates="+conf.KubeFeatureGates,
		)
	}

	var ports []internalversion.Port
	var volumes []internalversion.Volume
	volumes = append(volumes, conf.ExtraVolumes...)

	inContainer := conf.Image != ""
	if inContainer {
		kubeApiserverArgs = append(kubeApiserverArgs,
			"--etcd-servers=http://"+conf.EtcdAddress+":2379",
		)
	} else {
		kubeApiserverArgs = append(kubeApiserverArgs,
			"--etcd-servers=http://"+conf.EtcdAddress+":"+format.String(conf.EtcdPort),
		)
	}

	if conf.SecurePort {
		if conf.KubeAuthorization {
			kubeApiserverArgs = append(kubeApiserverArgs,
				"--authorization-mode=Node,RBAC",
			)
		}

		if inContainer {
			ports = []internalversion.Port{
				{
					HostPort: conf.Port,
					Port:     6443,
				},
			}
			volumes = append(volumes,
				internalversion.Volume{
					HostPath:  conf.CaCertPath,
					MountPath: "/etc/kubernetes/pki/ca.crt",
					ReadOnly:  true,
				},
				internalversion.Volume{
					HostPath:  conf.AdminCertPath,
					MountPath: "/etc/kubernetes/pki/admin.crt",
					ReadOnly:  true,
				},
				internalversion.Volume{
					HostPath:  conf.AdminKeyPath,
					MountPath: "/etc/kubernetes/pki/admin.key",
					ReadOnly:  true,
				},
			)
			kubeApiserverArgs = append(kubeApiserverArgs,
				"--bind-address="+publicAddress,
				"--secure-port=6443",
				"--tls-cert-file=/etc/kubernetes/pki/admin.crt",
				"--tls-private-key-file=/etc/kubernetes/pki/admin.key",
				"--client-ca-file=/etc/kubernetes/pki/ca.crt",
				"--service-account-key-file=/etc/kubernetes/pki/admin.key",
				"--service-account-signing-key-file=/etc/kubernetes/pki/admin.key",
				"--service-account-issuer=https://kubernetes.default.svc.cluster.local",
			)
		} else {
			kubeApiserverArgs = append(kubeApiserverArgs,
				"--bind-address="+conf.Address,
				"--secure-port="+format.String(conf.Port),
				"--tls-cert-file="+conf.AdminCertPath,
				"--tls-private-key-file="+conf.AdminKeyPath,
				"--client-ca-file="+conf.CaCertPath,
				"--service-account-key-file="+conf.AdminKeyPath,
				"--service-account-signing-key-file="+conf.AdminKeyPath,
				"--service-account-issuer=https://kubernetes.default.svc.cluster.local",
			)
		}
	} else {
		if inContainer {
			ports = []internalversion.Port{
				{
					HostPort: conf.Port,
					Port:     8080,
				},
			}

			kubeApiserverArgs = append(kubeApiserverArgs,
				"--insecure-bind-address="+publicAddress,
				"--insecure-port=8080",
			)
		} else {
			kubeApiserverArgs = append(kubeApiserverArgs,
				"--insecure-bind-address="+conf.Address,
				"--insecure-port="+format.String(conf.Port),
			)
		}
	}

	if conf.AuditPolicyPath != "" {
		if inContainer {
			volumes = append(volumes,
				internalversion.Volume{
					HostPath:  conf.AuditPolicyPath,
					MountPath: "/etc/kubernetes/audit-policy.yaml",
					ReadOnly:  true,
				},
				internalversion.Volume{
					HostPath:  conf.AuditLogPath,
					MountPath: "/var/log/kubernetes/audit/audit.log",
					ReadOnly:  false,
				},
			)
			kubeApiserverArgs = append(kubeApiserverArgs,
				"--audit-policy-file=/etc/kubernetes/audit-policy.yaml",
				"--audit-log-path=/var/log/kubernetes/audit/audit.log",
			)
		} else {
			kubeApiserverArgs = append(kubeApiserverArgs,
				"--audit-policy-file="+conf.AuditPolicyPath,
				"--audit-log-path="+conf.AuditLogPath,
			)
		}
	}

	if conf.Verbosity != int(log.LevelInfo) {
		kubeApiserverArgs = append(kubeApiserverArgs, "--v="+strconv.Itoa(conf.Verbosity))
	}

	return internalversion.Component{
		Name:    "kube-apiserver",
		Version: conf.Version.String(),
		Links: []string{
			"etcd",
		},
		Command: []string{"kube-apiserver"},
		Ports:   ports,
		Volumes: volumes,
		Args:    kubeApiserverArgs,
		Binary:  conf.Binary,
		Image:   conf.Image,
		WorkDir: conf.Workdir,
	}, nil
}
