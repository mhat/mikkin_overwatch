mikkin's overwatch
==================

Mikkin's a swell guy. For reasons we started naming services related to managing our development
environments after him. I'm not really sure why, it just happened and who am I to go against
such things.

With out diving into too much dotail here's the deal: Running Yammer requires running a lot of
services. We use a combination of Vagrant and Dazzel(internal) to make it easy to provision and
manage our development environments. We have consistent conventions so it's pretty easy to know
where to find logs, etc.

That said, there's a lot of services. A lot of logs. And when something isn't working it can be
pretty tedious to figure out where the problem is. It's easy enough to SSH into the VM and tail
all the log files but I wanted something easier ... 

... Overwatch is basically tail piped over a websocket and displayed in a browser.

It's by no means unique. I found a handful of similar services before I started. Normally I'd
just use one of those but I was also looking for a low risk project to experiment with GoLang
and so here we are. 


notes
=====

If you are using Ubuntu 12.04 with the 3.2.0 Kernel you need to make sure you're at patch 37 or
above. Lower patch levels have (or may have) a inotify leak: after subscribing to 128 files you
won't be able to subscribe to any more. It's cumulative and applies to commands like tail as 
well. So that's pretty neat. 

