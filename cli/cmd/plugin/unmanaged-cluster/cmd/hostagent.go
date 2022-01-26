// Copyright 2021 VMware Tanzu Community Edition contributors. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/lima-vm/lima/pkg/hostagent"
	"github.com/lima-vm/lima/pkg/hostagent/api/server"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const hostagentDesc = `
lima hostagent`

var HostagentCmd = &cobra.Command{
	Use:   "hostagent",
	Short: "lima hostagent",
	Long:  hostagentDesc,
	PreRunE: func(cmd *cobra.Command, args []string) (err error) {
		return nil
	},
	RunE: hostagentRun,
	PostRunE: func(cmd *cobra.Command, args []string) (err error) {
		return nil
	},
}

func init() {
	HostagentCmd.Flags().Bool("tty-disable", false, "Disable log stylization and emojis")
	HostagentCmd.Flags().StringP("pidfile", "p", "", "write pid to file")
	HostagentCmd.Flags().String("socket", "", "hostagent socket")
	HostagentCmd.Flags().String("nerdctl-archive", "", "local file path (not URL) of nerdctl-full-VERSION-linux-GOARCH.tar.gz")
}

// This command was copied from the lima-vm/lima hostagent command
// When running `makeVm`, it calls lima-vm/lima/pkg/start.Start()
// which uses os.Exec to call this command.
func hostagentRun(cmd *cobra.Command, args []string) error {
	pidfile, err := cmd.Flags().GetString("pidfile")
	if err != nil {
		return err
	}

	if pidfile != "" {
		if _, err := os.Stat(pidfile); !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("pidfile %q already exists", pidfile)
		}
		if err := os.WriteFile(pidfile, []byte(strconv.Itoa(os.Getpid())+"\n"), 0644); err != nil {
			return err
		}
		defer os.RemoveAll(pidfile)
	}
	socket, err := cmd.Flags().GetString("socket")
	if err != nil {
		return err
	}
	if socket == "" {
		return fmt.Errorf("socket must be specified (limactl version mismatch?)")
	}

	instName := args[0]

	sigintCh := make(chan os.Signal, 1)
	signal.Notify(sigintCh, os.Interrupt)

	var opts []hostagent.Opt
	nerdctlArchive, err := cmd.Flags().GetString("nerdctl-archive")
	if err != nil {
		return err
	}
	if nerdctlArchive != "" {
		opts = append(opts, hostagent.WithNerdctlArchive(nerdctlArchive))
	}
	ha, err := hostagent.New(instName, cmd.OutOrStdout(), sigintCh, opts...)
	if err != nil {
		return err
	}

	backend := &server.Backend{
		Agent: ha,
	}
	r := mux.NewRouter()
	server.AddRoutes(r, backend)
	srv := &http.Server{Handler: r}
	err = os.RemoveAll(socket)
	if err != nil {
		return err
	}
	l, err := net.Listen("unix", socket)
	if err != nil {
		return err
	}
	go func() {
		defer os.RemoveAll(socket)
		defer srv.Close()
		if serveErr := srv.Serve(l); serveErr != nil {
			logrus.WithError(serveErr).Warn("hostagent API server exited with an error")
		}
	}()
	return ha.Run(cmd.Context())
}
