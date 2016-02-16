package main

import (
	"fmt"
	"os"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var outputAsCSV bool

var listPeersCommand = &cobra.Command{
	Use:   "list-peers",
	Short: "show the active storage peers in the cluster",
	Run:   listPeersAction,
}

func init() {
	listPeersCommand.Flags().BoolVarP(&outputAsCSV, "csv", "", false, "output as csv instead")
}

func listPeersAction(cmd *cobra.Command, args []string) {
	mds := mustConnectToMDS()
	gmd, err := mds.GlobalMetadata()
	if err != nil {
		fmt.Fprintf(os.Stderr, "couldn't get global metadata: %s\n", err)
		os.Exit(1)
	}
	peers, err := mds.GetPeers()
	if err != nil {
		fmt.Fprintf(os.Stderr, "couldn't get peers: %s\n", err)
		os.Exit(1)
	}
	table := tablewriter.NewWriter(os.Stdout)
	if outputAsCSV {
		table.SetBorder(false)
		table.SetColumnSeparator(",")
	} else {
		table.SetHeader([]string{"Address", "UUID", "Size", "Used", "Updated", "Reb/Rep Data"})
	}
	rebalancing := false
	for _, x := range peers {
		if x.Address == "" {
			continue
		}
		table.Append([]string{
			x.Address,
			x.UUID,
			humanize.IBytes(x.TotalBlocks * gmd.BlockSize),
			humanize.IBytes(x.UsedBlocks * gmd.BlockSize),
			humanize.Time(time.Unix(0, x.LastSeen)),
			humanize.IBytes(x.LastRebalanceBlocks*gmd.BlockSize*uint64(time.Second)/uint64(x.LastSeen+1-x.LastRebalanceFinish)) + "/sec",
		})
		if x.Rebalancing {
			rebalancing = true
		}
	}
	table.Render()
	fmt.Printf("Balanced: %v\n", !rebalancing)
}