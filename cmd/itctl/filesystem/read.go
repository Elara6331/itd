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
	"io/ioutil"
	"os"
	"path/filepath"

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

		var tmpFile *os.File
		var path string
		var err error
		if args[1] == "-" {
			tmpFile, err = ioutil.TempFile("/tmp", "itctl.*")
			if err != nil {
				log.Fatal().Err(err).Msg("Error creating temporary file")
			}
			path = tmpFile.Name()
		} else {
			path, err = filepath.Abs(args[1])
			if err != nil {
				log.Fatal().Err(err).Msg("Error making absolute directory")
			}
		}

		client := viper.Get("client").(*api.Client)

		err = client.ReadFile(path, args[0])
		if err != nil {
			log.Fatal().Err(err).Msg("Error reading remote file")
		}

		if args[1] == "-" {
			io.Copy(os.Stdout, tmpFile)
			os.Stdout.WriteString("\n")
			tmpFile.Close()
			os.Remove(path)
		}
	},
}

func init() {
	filesystemCmd.AddCommand(readCmd)
}
