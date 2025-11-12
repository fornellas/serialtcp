package main

import (
	"github.com/spf13/cobra"

	"github.com/fornellas/slogxt/log"
)

var CompactCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start a TCP server connected to a serial port.",
	Args:  cobra.ExactArgs(1),
	Run: GetRunFn(func(cmd *cobra.Command, args []string) (err error) {

		ctx, logger := log.MustWithAttrs(
			cmd.Context(),
			// TODO cli flags
		)
		cmd.SetContext(ctx)
		logger.Info("Running")

		return
	}),
}

func init() {
	RootCmd.AddCommand(CompactCmd)
}
