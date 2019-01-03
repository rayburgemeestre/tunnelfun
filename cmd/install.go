package cmd

// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

import (
	"fmt"
	"github.com/spf13/cobra"
	"io/ioutil"
)

var (
	server bool
)

func init() {
	installCmd.Flags().StringVarP(&client, "client", "C", "", "install config for given client")
	installCmd.Flags().BoolVarP(&server, "server", "S", false, "install server config")
	rootCmd.AddCommand(installCmd)
}

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Run this to deploy systemd unit file for given client or server config.",
	Long: `This command will deploy tunnelfun.service to /etc/systemd/system.
		   You still need to systemctl daemon-reload, systemctl enable tunnelfun, systemctl start tunnelfun manually.
           Invoke using sudo tunnelfun install --server, sudo tunnelfun install --client <client>`,

	Run: func(cmd *cobra.Command, args []string) {
		if client == "" && !server {
			panic("No client or server specified, use -C <client> or -S.")
		}

		var unitContents string
		if client != "" {
			unitContents = fmt.Sprintf(`[Unit]
Description=Tunnelfun Client (%s)

[Service]
ExecStart=/bin/sh -c "while true; do tunnelfun --config %s client -C %s & sleep 10; done"
`, client, cfgFile, client)

		} else if server {
			unitContents = fmt.Sprintf(`[Unit]
Description=Tunnelfun Server (Reaper)

[Service]
ExecStart=/bin/sh -c "while true; do tunnelfun --config %s server & sleep 10; done"
`, cfgFile)
		}

		err := ioutil.WriteFile("/etc/systemd/system/tunnelfun.service", []byte(unitContents), 0644)
		if err != nil {
			panic(err)
		}
	},
}
