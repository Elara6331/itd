/*
 *	itd uses bluetooth low energy to communicate with InfiniTime devices
 *	Copyright (C) 2021 Arsen Musayelyan
 *
 *	This program is free software: you can redistribute it and/or modify
 *	it under the terms of the GNU General Public License as published by
 *	the Free Software Foundation, either version 3 of the License, or
 *	(at your option) any later version.
 *
 *	This program is distributed in the hope that it will be useful,
 *	but WITHOUT ANY WARRANTY; without even the implied warranty of
 *	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *	GNU General Public License for more details.
 *
 *	You should have received a copy of the GNU General Public License
 *	along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package root

import (
	"github.com/abiosoft/ishell"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "itctl",
	Short: "Control the itd daemon for InfiniTime smartwatches",
	Run: func(cmd *cobra.Command, args []string) {

		// Create new shell
		sh := ishell.New()
		sh.SetPrompt("itctl> ")

		// For every command in cobra
		for _, subCmd := range cmd.Commands() {
			// Add top level command to ishell
			sh.AddCmd(&ishell.Cmd{
				Name:     subCmd.Name(),
				Help:     subCmd.Short,
				Aliases:  subCmd.Aliases,
				LongHelp: subCmd.Long,
				Func: func(ctx *ishell.Context) {
					// Append name and arguments of command
					args := append([]string{ctx.Cmd.Name}, ctx.Args...)
					// Set root command arguments
					cmd.SetArgs(args)
					// Execute root command with new arguments
					cmd.Execute()
				},
			})
		}

		// Start shell
		sh.Run()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	RootCmd.CompletionOptions.DisableDefaultCmd = true
	cobra.CheckErr(RootCmd.Execute())
}

func init() {
	// Register flag for socket path
	RootCmd.Flags().StringP("socket-path", "s", "", "Path to itd socket")

	// Bind flag and environment variable to viper key
	viper.BindPFlag("sockPath", RootCmd.Flags().Lookup("socket-path"))
	viper.BindEnv("sockPath", "ITCTL_SOCKET_PATH")

	// Set default value for socket path
	viper.SetDefault("sockPath", "/tmp/itd/socket")
}
