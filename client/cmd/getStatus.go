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
		fmt.Printf("Job Status: %s \n", res.Status)
	},
}

func init() {
	rootCmd.AddCommand(getStatusCmd)

}
