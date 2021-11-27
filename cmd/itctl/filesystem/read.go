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
	"os"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.arsenm.dev/itd/api"
)

// readCmd represents the read command
var readCmd = &cobra.Command{
	Use:   `read <remote path> <local path | "-">`,
	Short: "Read a file from InfiniTime",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 2 {
			cmd.Usage()
			log.Fatal().Msg("Command read requires two arguments")
		}

		start := time.Now()
		client := viper.Get("client").(*api.Client)

		data, err := client.ReadFile(args[0])
		if err != nil {
			log.Fatal().Err(err).Msg("Error moving file or directory")
		}

		var suffix string
		var out *os.File
		if args[1] == "-" {
			out = os.Stdout
			suffix = "\n"
		} else {
			out, err = os.Create(args[1])
			if err != nil {
				log.Fatal().Err(err).Msg("Error opening local file")
			}
		}

		n, err := out.WriteString(data)
		if err != nil {
			log.Fatal().Err(err).Msg("Error writing to local file")
		}
		out.WriteString(suffix)
		
		log.Info().Msgf("Read %d bytes in %s", n, time.Since(start))
	},
}

func init() {
	filesystemCmd.AddCommand(readCmd)
}
