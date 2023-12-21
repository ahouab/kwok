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

// Package tidy provides a command to tidy the config file
package tidy

import (
	"context"

	"github.com/spf13/cobra"

	"sigs.k8s.io/kwok/pkg/config"
	"sigs.k8s.io/kwok/pkg/consts"
	"sigs.k8s.io/kwok/pkg/kwokctl/dryrun"
	"sigs.k8s.io/kwok/pkg/utils/path"
)

// NewCommand returns a new cobra.Command for config save
func NewCommand(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Args:  cobra.NoArgs,
		Use:   "tidy",
		Short: "Tidy the default config file. When combined with --config, it merges the specified configuration files into the default one.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runE(cmd.Context())
		},
	}
	return cmd
}

func runE(ctx context.Context) error {
	list := config.GetFromContext(ctx)
	p := path.Join(config.WorkDir, consts.ConfigName)
	if dryrun.DryRun {
		dryrun.PrintMessage("# Tidy the config file")
		return nil
	}
	err := config.Save(ctx, p, list)
	if err != nil {
		return err
	}
	return nil
}
