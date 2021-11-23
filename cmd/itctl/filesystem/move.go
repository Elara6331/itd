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

package filesystem

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.arsenm.dev/itd/api"
)

// heartCmd represents the heart command
var moveCmd = &cobra.Command{
	Use:     "move <old> <new>",
	Aliases: []string{"mv"},
	Short:   "Move a file or directory",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 2 {
			cmd.Usage()
			log.Fatal().Msg("Command move requires two arguments")
		}

		client := viper.Get("client").(*api.Client)

		err := client.Rename(args[0], args[1])
		if err != nil {
			log.Fatal().Err(err).Msg("Error moving file or directory")
		}
	},
}

func init() {
	filesystemCmd.AddCommand(moveCmd)
}
