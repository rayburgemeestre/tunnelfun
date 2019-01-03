[![Build Status](https://travis-ci.org/rayburgemeestre/tunnelfun.svg?branch=master)](https://travis-ci.org/rayburgemeestre/tunnelfun) [![MPL 2.0 License](https://img.shields.io/badge/license-MPL2.0-blue.svg)](http://veldstra.org/2016/12/09/you-should-choose-mpl2-for-your-opensource-project.html)

## About

This project is for my own convenience but might be useful for others.
I have all my devices connect to a central server and open a remote port
forward to their own SSH port.

Basically what you would do with the following SSH command:

    ssh -R 2000:localhost:22 root@mysever "while true; do uptime; sleep 2; done"

I used to have some bash scripts that would run to keep these tunnels alive,
with special features like restarting the tunnel if there hasn't been any new
output from the while loop.

## Why?

This way I can login to the server and reach all my devices like raspberry
pi's, NAS, even my own laptop, wherever they are running as long as they have
internet. Some are behind a few firewalls running at random locations, so not
easy to connect to. Also setting up those fancy tunnels that would survive
reboots etc., I wanted it to be as easy as possible and that is another reason
why I created this tool, it will generate a working systemd service for you.

## Implementation

It tries to be pro-active in making sure the tunnels are up & running. In
various occasions I would have stale processes that think they were connected
but really weren't because the WiFi network changed or whatever. For some
reason I couldn't rely on the SSH keep alive feature, it simply didn't seem to
work for me. Also, the server has the same problem, it would never kill off the
ssh processes that bind to port X, even though there was nobody actually
providing the service behind the tunnel anymore. I couldn't get this to work
with the right config settings, but by implementing this project at least I
know it should always work :)

Now with this tool on the central server a "reaper" runs, that kills stale ssh
processes that bind ports for tunnels that no longer work.
For example a tunnel running on port 2000, it will try to connect to the port
and search for the string "SSH" (which should be printed by the SSH server) If
it finds this string (within a given time frame) it leaves the tunnel alive,
otherwise it will kill it.
(We do not login to the tunnel because that would be a bit risky, one server
that can login to everywhere.)
Which will allow a new process to bind this port (probably the same tunnel,
which has a periodic retry mechanism).

The client will try to establish a tunnel every XX seconds, even if there is
already a tunnel "up & running". The way it works is there is the service will
try to create the tunnel constantly, if it succeeds, it will kill all other
existing processes, assuming they are malfunctioning.

## Usage

    tunnelfun - let's tunnel all the thingz!
    
    Usage:
      tunnelfun [command]
    
    Available Commands:
      client      Run this on a client to connect to tunnelserver.
      help        Help about any command
      install     Run this to deploy systemd unit file for given client or server config.
      server      Run this on the tunnel server, where all the tunnels connect to. Reaps dead ssh tunnels!
    
    Flags:
          --config string   config file (default is $HOME/tunnelfun.yaml)
      -h, --help            help for tunnelfun
    
    Use "tunnelfun [command] --help" for more information about a command.

Please see the example directory for a `tunnelfun.yaml` example.

* `tunnelfun server` - to run the server (reaper process) once.
* `tunnelfun client -C <name>` - to run the client (tunnel creator) once.
* `tunnelfun install -C <name>` - install systemd unit file that periodically runs the client tunnel command.
* `tunnelfun install -S` - install systemd unit file that periodically runs the server reaper command.

When deploying systemd services you still need to do the following:

* `systemctl daemon-reload`
* `systemctl enable tunnelfun`
* `systemctl start tunnelfun`
* `journalctl -fu tunnelfun` - to see what it's doing

In some cases, like on my Intel Edison device that runs Yocto linux, I needed to edit the generated systemd file a bit before I could enable it.
I've put this in the examples directory as well.

## Install

Download the binary and put it somewhere in your $PATH.

Put the example `tunnelfun.yaml` in your $HOME directory and edit it (obviously).

## Build

    go get https://github.com/rayburgemeestre/tunnelfun
    cd ~/go/src/github.com/rayburgemeestre/tunnelfun
    go build
    
For building arm (e.g. Raspberry Pi):

    env GOOS=linux GOARCH=arm GOARM=5 go build

For building 32 bit (e.g. Intel Edison board):

    env GOOS=linux GOARCH=386 go build
    
