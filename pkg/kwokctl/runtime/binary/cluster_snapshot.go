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

package binary

import (
	"context"
	"os"

	"sigs.k8s.io/kwok/pkg/kwokctl/runtime"
	"sigs.k8s.io/kwok/pkg/log"
	"sigs.k8s.io/kwok/pkg/utils/exec"
	"sigs.k8s.io/kwok/pkg/utils/file"
	"sigs.k8s.io/kwok/pkg/utils/format"
)

// SnapshotSave save the snapshot of cluster
func (c *Cluster) SnapshotSave(ctx context.Context, path string) error {
	config, err := c.Config(ctx)
	if err != nil {
		return err
	}
	conf := &config.Options

	etcdctlPath := c.GetBinPath("etcdctl" + conf.BinSuffix)

	err = file.DownloadWithCacheAndExtract(ctx, conf.CacheDir, conf.EtcdBinaryTar, etcdctlPath, "etcdctl"+conf.BinSuffix, 0750, conf.QuietPull, true)
	if err != nil {
		return err
	}

	err = exec.Exec(ctx, etcdctlPath, "snapshot", "save", path, "--endpoints=127.0.0.1:"+format.String(conf.EtcdPort))
	if err != nil {
		return err
	}

	return nil
}

// SnapshotRestore restore the snapshot of cluster
func (c *Cluster) SnapshotRestore(ctx context.Context, path string) error {
	config, err := c.Config(ctx)
	if err != nil {
		return err
	}
	conf := &config.Options

	etcdctlPath := c.GetBinPath("etcdctl" + conf.BinSuffix)

	err = file.DownloadWithCacheAndExtract(ctx, conf.CacheDir, conf.EtcdBinaryTar, etcdctlPath, "etcdctl"+conf.BinSuffix, 0750, conf.QuietPull, true)
	if err != nil {
		return err
	}

	logger := log.FromContext(ctx)

	err = c.StopComponent(ctx, "etcd")
	if err != nil {
		logger.Error("Failed to stop etcd", err)
	}
	defer func() {
		err = c.StartComponent(ctx, "etcd")
		if err != nil {
			logger.Error("Failed to start etcd", err)
		}
	}()

	etcdDataTmp := c.GetWorkdirPath("etcd-data")
	err = os.RemoveAll(etcdDataTmp)
	if err != nil {
		return err
	}

	err = exec.Exec(ctx, etcdctlPath, "snapshot", "restore", path, "--data-dir", etcdDataTmp)
	if err != nil {
		return err
	}

	etcdDataPath := c.GetWorkdirPath(runtime.EtcdDataDirName)
	err = os.RemoveAll(etcdDataPath)
	if err != nil {
		return err
	}
	err = os.Rename(etcdDataTmp, etcdDataPath)
	if err != nil {
		return err
	}
	return nil
}

// SnapshotSaveWithYAML save the snapshot of cluster
func (c *Cluster) SnapshotSaveWithYAML(ctx context.Context, path string, filters []string) error {
	err := c.Cluster.SnapshotSaveWithYAML(ctx, path, filters)
	if err != nil {
		return err
	}
	return nil
}

// SnapshotRestoreWithYAML restore the snapshot of cluster
func (c *Cluster) SnapshotRestoreWithYAML(ctx context.Context, path string, filters []string) error {
	logger := log.FromContext(ctx)
	err := c.StopComponent(ctx, "kube-controller-manager")
	if err != nil {
		logger.Error("Failed to stop kube-controller-manager", err)
	}
	defer func() {
		err = c.StartComponent(ctx, "kube-controller-manager")
		if err != nil {
			logger.Error("Failed to start kube-controller-manager", err)
		}
	}()

	err = c.Cluster.SnapshotRestoreWithYAML(ctx, path, filters)
	if err != nil {
		return err
	}
	return nil
}
