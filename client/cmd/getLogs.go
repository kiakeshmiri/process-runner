/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"

	prunner "github.com/kiakeshmiri/process-runner/api/protogen"
	"github.com/spf13/cobra"
)

// getLogsCmd represents the getLogs command
var getLogsCmd = &cobra.Command{
	Use:   "getLogs",
	Short: "Prints process logs",
	Long:  ``,
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
		opt := &prunner.GetLogsRequest{Uuid: uuid, Caller: cname}
		logStream, err := pr.GetLogs(ctx, opt)
		if err != nil {
			log.Fatal(err)
		}

		done := make(chan bool)

		go func() {
			for {
				resp, err := logStream.Recv()
				if err == io.EOF {
					done <- true //close(done)
					return
				}
				if err != nil {
					log.Fatalf("can not receive %v", err)
				}
				strLog := string(resp.Log)
				lines := strings.Split(strLog, "\n")
				for _, line := range lines {
					fmt.Println(line)
				}
			}
		}()

		<-done
	},
}

func init() {
	rootCmd.AddCommand(getLogsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// getLogsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// getLogsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
