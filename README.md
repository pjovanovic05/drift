# Drift

This is a tool for differentiating two running linux instances. It 
collects all installed packages, files, file acls, users, groups, etc. 
to compile an html report on each type of diff.

Optimally, one would have their infrastructure stored as code and 
easily deployable so the types of problems which this app is aimed at 
don't even occur. But in bare metal deployments, it's easy to take a 
shortcut and just change a configuration or install a package directly 
on the machine instead of following the procedures and documenting everything done.


## Getting started

The app consists of a single binary that runs as either the server 
or the client. The idea is that you run servers on two separate linux 
instances you wish to compare, and the client is ran on your machine 
from which you are collecting and reviewing the results.

A Vagrant definition for 3 VMs is also included to streamline the 
testing process.

The process:
* To compile the code, simply run `go build`, and the drift binary will appear after a short while. 
* Then you can start the Vagrant machines with `vagrant up`
* Start servers on target1 and target2 machines:

```bash
vagrant ssh target1 # do the same on target2
cd /vagrant
sudo ./drift -d -pass testpass
```
* start the diffing proces on the client machine:
```bash
vagrant ssh client
cd /vagrant
./drift -config test-conf.json
```
This will create several html files on the client (and in the 
directory on host where the procedure is run, if the /vagrant dir
on VMs is synchronized):
* acls-report.html - shows differences in acls
* files-report.html - shows differences in files
* packages-report.html - shows differences in installed packages
* users-report.html - shows differences in users on the systems.

These html files can be large and they depend on bootstrap to 
show the ui in a more dynamic fashion. The files can be quite large
so be prepared to let the browser render slowly.
Better report templates for this are on my TODO list.