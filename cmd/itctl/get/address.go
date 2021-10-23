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

package get

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.arsenm.dev/itd/internal/types"
)

// addressCmd represents the address command
var addressCmd = &cobra.Command{
	Use:     "address",
	Aliases: []string{"addr"},
	Short:   "Get InfiniTime's bluetooth address",
	Run: func(cmd *cobra.Command, args []string) {
		// Connect to itd UNIX socket
		conn, err := net.Dial("unix", viper.GetString("sockPath"))
		if err != nil {
			log.Fatal().Err(err).Msg("Error dialing socket. Is itd running?")
		}
		defer conn.Close()

		// Encode request into connection
		err = json.NewEncoder(conn).Encode(types.Request{
			Type: types.ReqTypeBtAddress,
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

		// Print returned value
		fmt.Println(res.Value)
	},
}

func init() {
	getCmd.AddCommand(addressCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// addressCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// addressCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
