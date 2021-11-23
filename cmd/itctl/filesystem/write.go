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
	"io"
	"os"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.arsenm.dev/itd/api"
)

// heartCmd represents the heart command
var writeCmd = &cobra.Command{
	Use:   `write <local path | "-"> <remote path>`,
	Short: "Write a file to InfiniTime",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 2 {
			cmd.Usage()
			log.Fatal().Msg("Command write requires two arguments")
		}

		start := time.Now()
		client := viper.Get("client").(*api.Client)

		var in *os.File
		if args[0] == "-" {
			in = os.Stdin
		} else {
			fl, err := os.Open(args[0])
			if err != nil {
				log.Fatal().Err(err).Msg("Error opening local file")
			}
			in = fl
		}

		data, err := io.ReadAll(in)
		if err != nil {
			log.Fatal().Err(err).Msg("Error moving file or directory")
		}

		err = client.WriteFile(args[1], string(data))
		if err != nil {
			log.Fatal().Err(err).Msg("Error writing to remote file")
		}

		log.Info().Msgf("Wrote %d bytes in %s", len(data), time.Since(start))
	},
}

func init() {
	filesystemCmd.AddCommand(writeCmd)
}
