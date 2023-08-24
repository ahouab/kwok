/*
Copyright 2023 The Kubernetes Authors.

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
	"sigs.k8s.io/kwok/pkg/apis/internalversion"
	"sigs.k8s.io/kwok/pkg/log"
	"sigs.k8s.io/kwok/pkg/utils/format"
	"sigs.k8s.io/kwok/pkg/utils/version"
)

// BuildMetricsServerComponentConfig is the configuration for building a metrics server component.
type BuildMetricsServerComponentConfig struct {
	Binary         string
	Image          string
	Version        version.Version
	Workdir        string
	BindAddress    string
	Port           uint32
	CaCertPath     string
	AdminCertPath  string
	AdminKeyPath   string
	KubeconfigPath string
	Verbosity      log.Level
	ExtraArgs      []internalversion.ExtraArgs
	ExtraVolumes   []internalversion.Volume
	ExtraEnvs      []internalversion.Env
}

// BuildMetricsServerComponent builds a metrics server component.
func BuildMetricsServerComponent(conf BuildMetricsServerComponentConfig) (component internalversion.Component, err error) {
	metricsServerArgs := []string{
		"--kubelet-preferred-address-types=InternalIP,ExternalIP,Hostname",
		"--kubelet-use-node-status-port",
		"--metric-resolution=15s",
	}
	metricsServerArgs = append(metricsServerArgs, extraArgsToStrings(conf.ExtraArgs)...)

	inContainer := conf.Image != ""
	user := ""
	var volumes []internalversion.Volume
	volumes = append(volumes, conf.ExtraVolumes...)
	var ports []internalversion.Port
	if inContainer {
		volumes = append(volumes,
			internalversion.Volume{
				HostPath:  conf.KubeconfigPath,
				MountPath: "/root/.kube/config",
				ReadOnly:  true,
			},
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

		metricsServerArgs = append(metricsServerArgs,
			"--bind-address="+conf.BindAddress,
			"--secure-port=4443",
			"--kubeconfig=/root/.kube/config",
			"--authentication-kubeconfig=/root/.kube/config",
			"--authorization-kubeconfig=/root/.kube/config",
			"--tls-cert-file=/etc/kubernetes/pki/admin.crt",
			"--tls-private-key-file=/etc/kubernetes/pki/admin.key",
		)
		if conf.Port != 0 {
			ports = []internalversion.Port{
				{
					HostPort: conf.Port,
					Port:     4443,
				},
			}
		}
		user = "root"
	} else {
		metricsServerArgs = append(metricsServerArgs,
			"--bind-address="+conf.BindAddress,
			"--secure-port="+format.String(conf.Port),
			"--kubeconfig="+conf.KubeconfigPath,
			"--authentication-kubeconfig="+conf.KubeconfigPath,
			"--authorization-kubeconfig="+conf.KubeconfigPath,
			"--tls-cert-file="+conf.AdminCertPath,
			"--tls-private-key-file="+conf.AdminKeyPath,
		)
	}

	if conf.Verbosity != log.LevelInfo {
		metricsServerArgs = append(metricsServerArgs, "--v="+format.String(log.ToKlogLevel(conf.Verbosity)))
	}

	envs := []internalversion.Env{}
	envs = append(envs, conf.ExtraEnvs...)

	return internalversion.Component{
		Name:    "metrics-server",
		Version: conf.Version.String(),
		Links: []string{
			"kwok-controller",
		},
		Command: []string{"/metrics-server"},
		User:    user,
		Ports:   ports,
		Volumes: volumes,
		Args:    metricsServerArgs,
		Binary:  conf.Binary,
		Image:   conf.Image,
		WorkDir: conf.Workdir,
		Envs:    envs,
	}, nil
}
