[![Build Status](https://travis-ci.org/rayburgemeestre/jirahours.svg?branch=master)](https://travis-ci.org/rayburgemeestre/jirahours) [![MPL 2.0 License](https://img.shields.io/badge/license-MPL2.0-blue.svg)](http://veldstra.org/2016/12/09/you-should-choose-mpl2-for-your-opensource-project.html)

## About

This project is for my own convenience but might be useful for others.
I have all my devices connect to a central server and open a remote port
forward to their own SSH port.

This way I can login to the server and reach all my devices like raspberry
pi's, NAS, even my own laptop, wherever they are running as long as they have
internet.

It tries to be pro-active in making sure the tunnels are up & running. In
various occasions I would have stale processes that think they were connected
but really weren't because the WiFi network changed or whatever. For some
reason I couldn't rely on the SSH keep alive feature, it simply didn't seem to
work for me, or at least it was difficult to debug.
The workaround I came up with is the following...

## How does it work

The server will actively check the tunnel ports, for example 2000, it will try
to SSH to it and execute the `uptime` command. If it works (within a given time
frame) it leaves the tunnel, otherwise it will kill it. Which will allow a new
process to bind this port (probably the same tunnel, which has a periodic retry
mechanism).

The client will try to establish a tunnel every XX seconds, even if there is
already a tunnel "up & running". The way it works is there is the service will
try to create the tunnel constantly, if it succeeds, it will kill all other
existing processes, assuming they are malfunctioning.

## Usage

TODO.. too lazy to write it now.

## Install

TODO.. too lazy to write it now.

