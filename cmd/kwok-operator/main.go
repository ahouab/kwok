/*
Copyright 2024 The Kubernetes Authors.

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

// Package main is the entry point for the kwok binary.
package main

import (
	"os"

	"github.com/spf13/pflag"

	"sigs.k8s.io/kwok/pkg/config"
	"sigs.k8s.io/kwok/pkg/log"
	"sigs.k8s.io/kwok/pkg/operator/cmd"
	"sigs.k8s.io/kwok/pkg/utils/signals"
)

func main() {
	flagset := pflag.NewFlagSet("global", pflag.ContinueOnError)
	flagset.ParseErrorsWhitelist.UnknownFlags = true
	flagset.Usage = func() {}

	ctx := signals.SetupSignalContext()
	ctx, logger := log.InitFlags(ctx, flagset)

	ctx, err := config.InitFlags(ctx, flagset)
	if err != nil {
		_, _ = os.Stderr.Write([]byte(flagset.FlagUsages()))
		logger.Error("Init config flags", err)
		os.Exit(1)
	}

	command := cmd.NewCommand(ctx)
	command.PersistentFlags().AddFlagSet(flagset)
	err = command.ExecuteContext(ctx)
	if err != nil {
		logger.Error("Execute exit", err)
		os.Exit(1)
	}
}
