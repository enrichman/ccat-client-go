package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"

	cat "github.com/enrichman/ccat-client-go"
	"github.com/spf13/cobra"
)

func NewChatCmd(catclient *cat.Client) *cobra.Command {
	chatCmd := &cobra.Command{
		Use: "chat",
		Run: func(cmd *cobra.Command, args []string) {
			in, out := make(chan string), make(chan string)

			go func() {
				fmt.Println("Say hi!")
				for {
					reader := bufio.NewReader(os.Stdin)
					line, _ := reader.ReadString('\n')
					line = strings.TrimSpace(line)
					in <- line
				}
			}()

			go func() {
				for {
					fmt.Println(<-out)
				}
			}()

			interrupt := make(chan os.Signal, 1)
			signal.Notify(interrupt, os.Interrupt)

			ctx, cancel := context.WithCancel(context.Background())
			go func() {
				<-interrupt
				log.Println("interrupt. Canceling")
				cancel()
			}()

			err := catclient.Chat.Chat(ctx, in, out)
			if err != nil {
				log.Println(err)
			}
		},
	}

	return chatCmd
}
