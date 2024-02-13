package cmd

import (
	"github.com/celestiaorg/celestia-node/nodebuilder"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
)

// Init constructs a CLI command to initialize Celestia Node of any type with the given flags.
func Init(fsets ...*flag.FlagSet) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialization for Celestia Node. Passed flags have persisted effect.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cfg := NodeConfig(ctx)
			path := StorePath(ctx)
			ring, err := nodebuilder.InitKeyring(&cfg, path)
			if err != nil {
				return err
			}
			return nodebuilder.Init(ring, cfg, path, NodeType(ctx))
		},
	}
	for _, set := range fsets {
		cmd.Flags().AddFlagSet(set)
	}
	return cmd
}
