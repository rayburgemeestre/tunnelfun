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
easy to connect to.

## Implementation

It tries to be pro-active in making sure the tunnels are up & running. In
various occasions I would have stale processes that think they were connected
but really weren't because the WiFi network changed or whatever. For some
reason I couldn't rely on the SSH keep alive feature, it simply didn't seem to
work for me.

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

TODO.. too lazy to write it now.

## Install

TODO.. too lazy to write it now.

