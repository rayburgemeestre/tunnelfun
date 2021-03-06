package cmd

// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io"
	"log"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

func init() {
	rootCmd.AddCommand(serverCmd)
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Run this on the tunnel server, where all the tunnels connect to. Reaps dead ssh tunnels!",
	Long: `The server command runs once, so you run it in a loop.
			It checks lsof -i -P -n and finds all processes that LISTEN on ports.
			When the ports are in the list of monitoring ports it will check if it correctly tunnels SSH,
			if it doesn't it will kill the process.'`,

	Run: func(cmd *cobra.Command, args []string) {
		m := viper.GetStringSlice("server.tunnelports")
		timeoutResponse := viper.GetInt("server.timeout.response")
		monitoring := map[int]bool{}
		for _, k := range m {
			p, err := strconv.Atoi(k)
			if err != nil {
				panic(err)
			}
			monitoring[p] = true
		}

		found := map[int]int{}

		out, err := exec.Command("lsof", "-i", "-P", "-n").Output()
		if err != nil {
			log.Fatal(err)
		}
		lines := strings.Split(string(out), "\n")
		for _, line := range lines {
			chunks := strings.Fields(line)
			if len(chunks) < 9 {
				continue
			}
			if chunks[7] != "TCP" {
				continue
			}
			if !strings.Contains(chunks[9], "LISTEN") {
				continue
			}

			spec := strings.Split(chunks[8], ":")
			if len(spec) < 2 {
				continue
			}
			port, err := strconv.Atoi(spec[1])
			if err != nil {
				continue
			}
			if _, exists := monitoring[port]; exists {
				pid, err := strconv.Atoi(chunks[1])
				if err != nil {
					panic(err)
				}
				found[port] = pid
			}
		}

		wg := sync.WaitGroup{}

		for port, pid := range found {
			wg.Add(1)
			go func(port int, pid int) {
				defer wg.Done()

				channel1 := make(chan string, 1)
				go func() {
					conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", port))
					if err != nil {
						fmt.Println("dial error:", err)
						return
					}
					defer conn.Close()

					buf := make([]byte, 0, 4096)
					tmp := make([]byte, 256)
					for {
						n, err := conn.Read(tmp)
						if err != nil {
							if err != io.EOF {
								fmt.Println("read error:", err)
							}
							break
						}
						buf = append(buf, tmp[:n]...)

						if strings.Contains(string(buf), "SSH") {
							channel1 <- string(buf)
							return
						}

					}
					channel1 <- string(buf)
				}()
				select {
				case res := <-channel1:
					if strings.Contains(res, "SSH") {
						fmt.Println("port", port, "- ok")
					} else {
						fmt.Println("port", port, "- killing pid", pid, "due to unexpected server response")
						err := syscall.Kill(pid, syscall.SYS_KILL)
						if err != nil {
							fmt.Println(err)
						}
					}
				case <-time.After(time.Duration(timeoutResponse) * time.Second):
					fmt.Println("port", port, "- killing pid", pid, "due to timeout")
					err := syscall.Kill(pid, syscall.SYS_KILL)
					if err != nil {
						fmt.Println(err)
					}
				}
			}(port, pid)
		}
		wg.Wait()
		fmt.Println("done.")
	},
}
