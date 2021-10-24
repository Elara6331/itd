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

package firmware

import (
	"github.com/cheggaaa/pb/v3"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.arsenm.dev/itd/api"
	"go.arsenm.dev/itd/internal/types"
)

type DFUProgress struct {
	Received int64 `mapstructure:"recvd"`
	Total    int64 `mapstructure:"total"`
}

// upgradeCmd represents the upgrade command
var upgradeCmd = &cobra.Command{
	Use:     "upgrade",
	Short:   "Upgrade InfiniTime firmware using files or archive",
	Aliases: []string{"upg"},
	Run: func(cmd *cobra.Command, args []string) {
		client := viper.Get("client").(*api.Client)

		var upgType api.UpgradeType
		var files []string
		// Get relevant data struct
		if viper.GetString("archive") != "" {
			// Get archive data struct
			upgType = types.UpgradeTypeArchive
			files = []string{viper.GetString("archive")}
		} else if viper.GetString("initPkt") != "" && viper.GetString("firmware") != "" {
			// Get files data struct
			upgType = types.UpgradeTypeFiles
			files = []string{viper.GetString("initPkt"), viper.GetString("firmware")}
		} else {
			cmd.Usage()
			log.Warn().Msg("Upgrade command requires either archive or init packet and firmware.")
			return
		}

		progress, err := client.FirmwareUpgrade(upgType, files...)
		if err != nil {
			log.Fatal().Err(err).Msg("Error initiating DFU")
		}

		// Create progress bar template
		barTmpl := `{{counters . }} B {{bar . "|" "-" (cycle .) " " "|"}} {{percent . }} {{rtime . "%s"}}`
		// Start full bar at 0 total
		bar := pb.ProgressBarTemplate(barTmpl).Start(0)
		// Create new scanner of connection
		for event := range progress {
			// Set total bytes in progress bar
			bar.SetTotal(event.Total)
			// Set amount of bytes received in progress bar
			bar.SetCurrent(event.Received)
			// If transfer finished, break
			if event.Received == event.Total {
				break
			}
		}
		// Finish progress bar
		bar.Finish()
	},
}

func init() {
	firmwareCmd.AddCommand(upgradeCmd)

	// Register flags
	upgradeCmd.Flags().StringP("archive", "a", "", "Path to firmware archive")
	upgradeCmd.Flags().StringP("init-pkt", "i", "", "Path to init packet (.dat file)")
	upgradeCmd.Flags().StringP("firmware", "f", "", "Path to firmware image (.bin file)")

	// Bind flags to viper keys
	viper.BindPFlag("archive", upgradeCmd.Flags().Lookup("archive"))
	viper.BindPFlag("initPkt", upgradeCmd.Flags().Lookup("init-pkt"))
	viper.BindPFlag("firmware", upgradeCmd.Flags().Lookup("firmware"))
}
