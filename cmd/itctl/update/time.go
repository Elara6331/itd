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

package update

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.arsenm.dev/itd/api"
)

// weatherCmd represents the time command
var weatherCmd = &cobra.Command{
	Use:   `weather`,
	Short: "Force an immediate update of weather data",
	Run: func(cmd *cobra.Command, args []string) {
		client := viper.Get("client").(*api.Client)

		err := client.UpdateWeather()
		if err != nil {
			log.Fatal().Err(err).Msg("Error updating weather")
		}
	},
}

func init() {
	updateCmd.AddCommand(weatherCmd)
}
