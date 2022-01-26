// Copyright 2021 VMware Tanzu Community Edition contributors. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/spf13/cobra"
	logger "github.com/vmware-tanzu/community-edition/cli/cmd/plugin/unmanaged-cluster/log"

	networks "github.com/lima-vm/lima/pkg/networks/reconcile"
	"github.com/lima-vm/lima/pkg/start"
	"github.com/lima-vm/lima/pkg/store"
)

const makeVmDesc = `
List known unmanaged clusters. This makeVm is produced by locating clusters saved to
$HOME/.config/tanzu/tkg/unmanaged`

// MakeVmCmd makes a new VM from the ... library
var MakeVmCmd = &cobra.Command{
	Use:   "makeVm",
	Short: "makes a new vm from the ... library",
	Long:  makeVmDesc,
	PreRunE: func(cmd *cobra.Command, args []string) (err error) {
		return nil
	},
	RunE: makeVm,
	PostRunE: func(cmd *cobra.Command, args []string) (err error) {
		return nil
	},
}

func init() {
	MakeVmCmd.Flags().Bool("tty-disable", false, "Disable log stylization and emojis")
}

// makeVm outputs a makeVm of all unmanaged clusters on the system.
func makeVm(cmd *cobra.Command, args []string) error {
	log := logger.NewLogger(TtySetting(cmd.Flags()), 0)

	// This is an experimental default lima instance configuration.
	// using the libraries provided by lima-vm/lima/pkg

	// In order for this experiment to work correctly, you _must_ have the following:
	// - the full path to ~/.lima/{name-of-instance} directory
	// - the default instance yaml file must be in that directory:
	//    * https://github.com/lima-vm/lima/blob/master/pkg/limayaml/default.yaml
	//          * Note: this yaml file gets loaded in at compile time when building `lima-vm/lima`
	//            Ref: https://github.com/lima-vm/lima/blob/f4863764aed67e77874a12c60b016dbda21b7879/pkg/limayaml/template.go#L7-L8
	// - the above yaml file in the named directory must be called "lima.yaml"
	// - install qemu and the qemu-utils packages
	//
	// - `make all` from the lima github project and `cp` the linux host binaries
	//   to the exact location where you are executing this command
	//    * These binaries are baked images for the machines and the expected path they are loaded from
	//      is using the location of the executing binary. In other words, during lima's install,
	//      these binaries are created and placed _next_ to the limactl binary.

	inst := &store.Instance{
		Name: "unmanaged-default",
		Dir:  "/home/jmcb/.lima/unmanaged-default",
	}

	ctx := cmd.Context()
	err := networks.Reconcile(ctx, inst.Name)
	if err != nil {
		log.Errorf("Error reconciling networks: %s\n", err.Error())
		return nil
	}

	err = start.Start(ctx, inst)
	if err != nil {
		log.Errorf("Error starting lima default machine: %s\n", err.Error())
		return nil
	}

	log.Event(logger.EnvelopeEmoji, "All done!")
	return nil
}
