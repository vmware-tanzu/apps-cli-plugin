// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package plugin

import (
	"github.com/spf13/cobra"
)

func newPostInstallCmd(desc *PluginDescriptor) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "post-install",
		Short:        "Run post install configuration for a plugin",
		Long:         "Run post install configuration for a plugin",
		Hidden:       true,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// invoke postInstall for the plugin
			return desc.PostInstallHook()
		},
	}

	return cmd
}
