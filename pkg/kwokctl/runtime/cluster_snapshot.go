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

package runtime

import (
	"context"
	"io"
	"os"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/api/meta"

	"sigs.k8s.io/kwok/pkg/kwokctl/dryrun"
	"sigs.k8s.io/kwok/pkg/kwokctl/snapshot"
	"sigs.k8s.io/kwok/pkg/log"
	"sigs.k8s.io/kwok/pkg/utils/client"
	"sigs.k8s.io/kwok/pkg/utils/yaml"
)

// SnapshotSaveWithYAML save the snapshot of cluster
func (c *Cluster) SnapshotSaveWithYAML(ctx context.Context, path string, conf SnapshotSaveWithYAMLConfig) error {
	if c.IsDryRun() {
		dryrun.PrintMessage("kubectl get %s -o yaml >%s", strings.Join(conf.Filters, ","), path)
		return nil
	}

	clientset, err := c.GetClientset(ctx)
	if err != nil {
		return err
	}

	restMapper, err := clientset.ToRESTMapper()
	if err != nil {
		return err
	}

	logger := log.FromContext(ctx)

	filters, errs := client.MappingForResources(restMapper, conf.Filters)
	if len(errs) > 0 {
		for _, err := range errs {
			logger.Error("failed to get mapping", err)
		}
	}

	f, err := c.OpenFile(path)
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
	}()

	saver, err := snapshot.NewSaver(clientset, snapshot.SaveConfig{
		Filters: filters,
	})
	if err != nil {
		return err
	}

	var w io.Writer = f
	if conf.Relative {
		startTime := time.Now()
		w = snapshot.NewWriteHook(w, func(b []byte) []byte {
			return snapshot.ReplaceTimeToRelative(startTime, b)
		})
	}

	encoder := yaml.NewEncoder(w)

	var tracks map[*meta.RESTMapping]*snapshot.TrackData
	if conf.Record {
		tracks = make(map[*meta.RESTMapping]*snapshot.TrackData)
	}
	err = saver.Save(ctx, encoder, tracks)
	if err != nil {
		return err
	}

	if conf.Record {
		err = saver.Record(ctx, encoder, tracks)
		if err != nil {
			return err
		}
	}

	return nil
}

// SnapshotRestoreWithYAML restore the snapshot of cluster
func (c *Cluster) SnapshotRestoreWithYAML(ctx context.Context, path string, conf SnapshotRestoreWithYAMLConfig) error {
	if c.IsDryRun() {
		dryrun.PrintMessage("kubectl create -f %s", path)
		return nil
	}

	clientset, err := c.GetClientset(ctx)
	if err != nil {
		return err
	}

	restMapper, err := clientset.ToRESTMapper()
	if err != nil {
		return err
	}

	f, err := os.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
	}()

	logger := log.FromContext(ctx)

	filters, errs := client.MappingForResources(restMapper, conf.Filters)
	if len(errs) > 0 {
		for _, err := range errs {
			logger.Error("failed to get mapping", err)
		}
	}

	loader, err := snapshot.NewLoader(clientset, snapshot.LoadConfig{
		NoFilers: len(filters) == 0,
		Filters:  filters,
	})
	if err != nil {
		return err
	}

	var r io.Reader = f

	if conf.Relative {
		startTime := time.Now()
		r = snapshot.NewReadHook(r, func(b []byte) []byte {
			return snapshot.RevertTimeFromRelative(startTime, b)
		})
	}

	decoder := yaml.NewDecoder(r)
	err = loader.Load(ctx, decoder)
	if err != nil {
		return err
	}

	if conf.Replay {
		err = loader.Replay(ctx, decoder)
		if err != nil {
			return err
		}
	}

	return nil
}
