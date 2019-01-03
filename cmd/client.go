package cmd

// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// This file contains some code based on Davide Dal Farra's code.
// His header file was:

/*
Go-Language implementation of an SSH Reverse Tunnel, the equivalent of below SSH command:
   ssh -R 8080:127.0.0.1:8080 operatore@146.148.22.123
which opens a tunnel between the two endpoints and permit to exchange information on this direction:
   server:8080 -----> client:8080
   once authenticated a process on the SSH server can interact with the service answering to port 8080 of the client
   without any NAT rule via firewall
Copyright 2017, Davide Dal Farra
MIT License, http://www.opensource.org/licenses/mit-license.php
*/

import (
	"bytes"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	client string
)

func init() {
	clientCmd.Flags().StringVarP(&client, "client", "C", "", "client name from config to use")
	rootCmd.AddCommand(clientCmd)
}

var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "Run this on a client to connect to tunnelserver.",
	Long: `The command runs, logs in to the server and tries to open a remote port redirect there.
			Anyone connected to that server should be able to ssh to the clients ssh server via this port.
			When it succesfully sets this tunnel up, it makes sure to kill all other clients, assuming they
			have become zombies.'`,

	Run: func(cmd *cobra.Command, args []string) {
		if client == "" {
			panic("No client config specified, use -C")
		}
		wg := sync.WaitGroup{}
		user := viper.GetString("clients." + client + ".user")
		publicKey := viper.GetString("clients." + client + ".publickey")
		localService := Endpoint{viper.GetString("clients." + client + ".localservice.host"), viper.GetInt("clients." + client + ".localservice.port")}
		remoteEndpoint := Endpoint{viper.GetString("clients." + client + ".remoteendpoint.host"), viper.GetInt("clients." + client + ".remoteendpoint.port")}
		serverEndpoint := Endpoint{viper.GetString("server.host"), viper.GetInt("server.port")}
		clientConnectTimeout := viper.GetInt("client.timeout.connect")

		// refer to https://godoc.org/golang.org/x/crypto/ssh for other authentication types
		sshConfig := &ssh.ClientConfig{
			User: user,
			Auth: []ssh.AuthMethod{
				publicKeyFile(publicKey, ""),
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			Timeout:         time.Duration(clientConnectTimeout) * time.Second,
		}

		serverConn, err := ssh.Dial("tcp", serverEndpoint.String(), sshConfig)
		if err != nil {
			log.Fatalln(fmt.Printf("Dial INTO remote server error: %s", err))
		}

		listener, err := serverConn.Listen("tcp", remoteEndpoint.String())
		if err != nil {
			log.Fatalln(fmt.Printf("Listen open port ON remote server error: %s", err))
		}
		defer listener.Close()

		fmt.Println("Port forwarding succeeded")

		// Kill other possibly stale instances of this program
		err = filepath.Walk("/proc", findAndKillProcess)
		if err != nil {
			if err == io.EOF {
				// Not an error, just a signal when we are done
				err = nil
			} else {
				log.Fatal(err)
			}
		}

		// handle incoming connections on reverse forwarded tunnel
		defer wg.Wait()
		for {
			// Open a (local) connection to localservice whose content will be forwarded so serverEndpoint
			local, err := net.Dial("tcp", localService.String())
			if err != nil {
				log.Fatalln(fmt.Printf("Dial INTO local service error: %s", err))
			}

			client, err := listener.Accept()
			fmt.Println("Accepted client..")
			if err != nil {
				log.Fatalln(err)
			}

			wg.Add(1)
			go handleClient(&wg, client, local)
		}
	},
}

type Endpoint struct {
	Host string
	Port int
}

func (endpoint *Endpoint) String() string {
	return fmt.Sprintf("%s:%d", endpoint.Host, endpoint.Port)
}

// From https://sosedoff.com/2015/05/25/ssh-port-forwarding-with-go.html
// Handle local client connections and tunnel data to the remote server
// Will use io.Copy - http://golang.org/pkg/io/#Copy
func handleClient(wg *sync.WaitGroup, client net.Conn, remote net.Conn) {
	defer client.Close()
	defer wg.Done()
	chDone := make(chan bool)

	// Start remote -> local data transfer
	go func() {
		_, err := io.Copy(client, remote)
		if err != nil {
			log.Println(fmt.Sprintf("error while copy remote->local: %s", err))
		}
		chDone <- true
	}()

	// Start local -> remote data transfer
	go func() {
		_, err := io.Copy(remote, client)
		if err != nil {
			log.Println(fmt.Sprintf("error while copy local->remote: %s", err))
		}
		chDone <- true
	}()

	<-chDone
}

func publicKeyFile(file string, password string) ssh.AuthMethod {
	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatalln(fmt.Sprintf("Cannot read SSH public key file %s", file))
		return nil
	}

	var key ssh.Signer
	if len(password) != 0 {
		key, err = ssh.ParsePrivateKeyWithPassphrase(buffer, []byte(password))
	} else {
		key, err = ssh.ParsePrivateKey(buffer)
	}
	if err != nil {
		log.Fatalln(fmt.Sprintf("Cannot parse SSH public key file %s", file))
		return nil
	}
	return ssh.PublicKeys(key)
}

// source https://stackoverflow.com/questions/41060457/golang-kill-process-by-name

func findAndKillProcess(path string, info os.FileInfo, err error) error {
	if err != nil {
		return nil
	}
	if strings.Count(path, "/") == 3 {
		if strings.Contains(path, "/status") {
			pid, err := strconv.Atoi(path[6:strings.LastIndex(path, "/")])
			if err != nil {
				log.Println(err)
				return nil
			}
			f, err := ioutil.ReadFile(path)
			if err != nil {
				log.Println(err)
				return nil
			}
			name := string(f[6:bytes.IndexByte(f, '\n')])
			if name == os.Args[0] && pid != os.Getpid() {
				fmt.Printf("PID: %d, Name: %s will be killed.\n", pid, name)
				proc, err := os.FindProcess(pid)
				if err != nil {
					log.Println(err)
				}
				err = proc.Kill()
				if err != nil {
					panic(err)
				}
				return io.EOF
			}
		}
	}
	return nil
}
