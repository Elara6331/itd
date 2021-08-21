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

package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"go.arsenm.dev/itd/internal/types"
)

// heartCmd represents the heart command
var heartCmd = &cobra.Command{
	Use:   "heart",
	Short: "Get heart rate from InfiniTime",
	Run: func(cmd *cobra.Command, args []string) {
		// Connect to itd UNIX socket
		conn, err := net.Dial("unix", SockPath)
		if err != nil {
			log.Fatal().Err(err).Msg("Error dialing socket. Is itd running?")
		}
		defer conn.Close()

		// Encode request into connection
		err = json.NewEncoder(conn).Encode(types.Request{
			Type: ReqTypeHeartRate,
		})
		if err != nil {
			log.Fatal().Err(err).Msg("Error making request")
		}

		// Read one line from connection
		line, _, err := bufio.NewReader(conn).ReadLine()
		if err != nil {
			log.Fatal().Err(err).Msg("Error reading line from connection")
		}

		var res types.Response
		// Decode line into response
		err = json.Unmarshal(line, &res)
		if err != nil {
			log.Fatal().Err(err).Msg("Error decoding JSON data")
		}

		if res.Error {
			log.Fatal().Msg(res.Message)
		}

		// Print returned BPM
		fmt.Printf("%d BPM\n", int(res.Value.(float64)))
	},
}

func init() {
	getCmd.AddCommand(heartCmd)
}
