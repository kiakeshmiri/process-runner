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

// stopJobCmd represents the stopJob command
var stopJobCmd = &cobra.Command{
	Use:   "stopJob",
	Short: "Stops the running job by passing uuid",
	Long:  `TBD`,
	Run: func(cmd *cobra.Command, args []string) {
		pr, cname, err := NewClient()
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}

		ctx := context.WithoutCancel(context.Background())

		uuid := args[0]

		if err != nil {
			log.Fatal(err)
		}

		opt := &prunner.StopProcessRequest{Uuid: uuid, Caller: cname}

		res, err := pr.Stop(ctx, opt)

		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(res)
	},
}

func init() {
	rootCmd.AddCommand(stopJobCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// stopJobCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// stopJobCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
