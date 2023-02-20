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
	"golang.org/x/exp/slog"
	"sigs.k8s.io/kwok/pkg/apis/internalversion"
	"sigs.k8s.io/kwok/pkg/utils/format"
	"sigs.k8s.io/kwok/pkg/utils/version"
)

// BuildPrometheusComponentConfig is the configuration for building a prometheus component.
type BuildPrometheusComponentConfig struct {
	Binary        string
	Image         string
	Version       version.Version
	Workdir       string
	Address       string
	Port          uint32
	ConfigPath    string
	AdminCertPath string
	AdminKeyPath  string
	Verbosity     int
}

// BuildPrometheusComponent builds a prometheus component.
func BuildPrometheusComponent(conf BuildPrometheusComponentConfig) (component internalversion.Component, err error) {
	if conf.Address == "" {
		conf.Address = publicAddress
	}

	prometheusArgs := []string{}

	inContainer := conf.Image != ""
	var volumes []internalversion.Volume
	var ports []internalversion.Port
	if inContainer {
		volumes = append(volumes,
			internalversion.Volume{
				HostPath:  conf.ConfigPath,
				MountPath: "/etc/prometheus/prometheus.yaml",
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
		ports = []internalversion.Port{
			{
				HostPort: conf.Port,
				Port:     9090,
			},
		}
		prometheusArgs = append(prometheusArgs,
			"--config.file=/etc/prometheus/prometheus.yaml",
			"--web.listen-address="+publicAddress+":9090",
		)
	} else {
		prometheusArgs = append(prometheusArgs,
			"--config.file="+conf.ConfigPath,
			"--web.listen-address="+conf.Address+":"+format.String(conf.Port),
		)
	}

	if conf.Verbosity != int(slog.InfoLevel) {
		prometheusArgs = append(prometheusArgs, "--log.level="+format.StringifyLevel(conf.Verbosity))
	}

	return internalversion.Component{
		Name:    "prometheus",
		Version: conf.Version.String(),
		Links: []string{
			"etcd",
			"kube-apiserver",
			"kube-controller-manager",
			"kube-scheduler",
			"kwok-controller",
		},
		Command: []string{"prometheus"},
		Ports:   ports,
		Volumes: volumes,
		Args:    prometheusArgs,
		Binary:  conf.Binary,
		Image:   conf.Image,
		WorkDir: conf.Workdir,
	}, nil
}
