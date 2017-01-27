# blackbox

*forward files on to syslog*

![Black Box Flight Recorder](http://i.imgur.com/sCSNdzU.jpg)

## About

Applications often provide only a limited ability to log to syslog and often
don't log in a consistent format. I also think that syslog is an operational
concern and the application should not know about where it is logging. Blackbox
is an experiment to decouple syslogging from an application without messing
about with syslog configuration (which is tricky on BOSH VMs).

Blackbox will tail all files in sub-directories of a specified `source_dir`, and forward any new lines to a syslog server.

## Usage

```
blackbox -config config.yml
```

The configuration file schema is as follows:

``` yaml
hostname: this-host

syslog:
  destination:
    transport: udp
    address: logs.example.com:1234

  source_dir: /path/to/log-dir
```

Consider the case where `log-dir` has the following structure:

```
/path/to/log-dir
|-- app1
|   |-- stdout.log
|   `-- stderr.log
`-- app2
    |-- foo.log
    `-- bar.log
```

Any new lines written to `app1/stdout.log` and `app1/stderr.log` get sent to syslog tagged as `app1`, while new lines written to `app2/foo.log` and `app2/bar.log` get sent to syslog tagged as `app2`.

Currently the priority and facility are hardcoded to `INFO` and `user`.

## Installation

```
go get -u github.com/concourse/blackbox/cmd/blackbox
```
