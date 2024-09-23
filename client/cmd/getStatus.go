/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"log"

	prunner "github.com/kiakeshmiri/process-runner/api/protogen"
	"github.com/spf13/cobra"
)

// getStatusCmd represents the getStatus command
var getStatusCmd = &cobra.Command{
	Use:   "getStatus",
	Short: "provides status of the process",
	Long:  `TBD`,
	Run: func(cmd *cobra.Command, args []string) {
		pr, user, err := NewClient()
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}

		ctx := context.WithoutCancel(context.Background())

		uuid := args[0]

		if err != nil {
			log.Fatal(err)
		}

		req := &prunner.GetStatusRequest{Uuid: uuid, Caller: user}

		res, err := pr.GetStatus(ctx, req)

		if err != nil {
			log.Fatal(err)
		}
		var status string

		switch res.Status {
		case prunner.Status_RUNNING:
			status = "Running"
		case prunner.Status_STOPPED:
			status = "Stopped"
		case prunner.Status_COMPLETED:
			status = "Completed"
		case prunner.Status_EXITEDWITHERROR:
			status = "Exited with error"

		}
		fmt.Printf("Job Status: %s \n", status)
	},
}

func init() {
	rootCmd.AddCommand(getStatusCmd)

}
