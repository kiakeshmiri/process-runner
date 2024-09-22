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

// startJobCmd represents the startJob command
var startJobCmd = &cobra.Command{
	Use:   "startJob",
	Short: "Starts a job on linux machine",
	Long:  `TBD`,
	Run: func(cmd *cobra.Command, args []string) {
		pr, cname, err := NewClient()
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}

		// Contact the server and print out its response.
		ctx := context.WithoutCancel(context.Background())

		job := args[0]
		jobArgs := args[1:]

		opt := &prunner.StartProcessRequest{Job: job, Args: jobArgs, Caller: cname}

		res, err := pr.Start(ctx, opt)

		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(res)
	},
}

func init() {
	rootCmd.AddCommand(startJobCmd)
}
