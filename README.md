# Drift

This is a tool for differentiating two running linux instances. It collects all installed packages, files, file acls, users, groups, etc. to compile an html report on each type of diff.

Optimally, one would have their infrastructure stored as code and easily deployable
so the types of problems which this app is aimed at don't even occur. But in 
bare metal deployments, it's easy to take a shortcut and just change a configuration
or install a package directly on the machine instead of doing everything by the book.

## Getting started

The app consists of a single binary that runs as either the server or the client. The idea
is that you run servers on two separate linux instances you wish to compare, and
the client is ran on your machine from which you are collecting and reviewing the results.

```bash
bla blabla
```