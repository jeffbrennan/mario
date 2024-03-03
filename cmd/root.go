package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "mario",
	Short: "Mario - an ADF monitoring tool",
	Run: func(cmd *cobra.Command, args []string) {
		startPersistentProcess()
	},
}

func startPersistentProcess() {
	var wg sync.WaitGroup
	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, syscall.SIGINT, syscall.SIGTERM)

	wg.Add(1)
	go func() {
		defer wg.Done()
		reader := bufio.NewReader(os.Stdin)
		for {
			select {
			case <-stopCh:
				fmt.Println("Stopping Mario...")
				return
			default:
				fmt.Print("mario>")
				userInput, err := reader.ReadString('\n')
				if err != nil {
					fmt.Println("Error reading input:", err)
					continue
				}

				userInput = strings.TrimSpace(userInput)
				parts := strings.Fields(userInput)
				if len(parts) == 0 {
					continue
				}

				command := parts[0]
				args := parts[1:]

				switch command {
				case "summarize":
					summarizeCmd.ParseFlags(args)
					summarizeCmd.Run(summarizeCmd, nil)

				case "compare":
					compareCmd.ParseFlags(args)
					compareCmd.Run(compareCmd, nil)

				case "exit":
					exitCmd.Run(exitCmd, nil)

				default:
					fmt.Println("Unknown command. Try again.")
				}
			}
		}
	}()

	<-stopCh
	wg.Wait()
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
