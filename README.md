# bucketbench
Bucketbench is a simple framework for running defined sequences
of lifecycle container operations against three different container
engines today: the full Docker engine, OCI's runc, and containerd.

Given a **bucket** is a physical type of container, the name is my attempt to
get away from calling it "dockerbench," given it runs against other
container engines as well. All attempts to come up with a more interesting
name failed before initial release. Suggestions welcome!

## Background
This project came about via some performance comparison work happening
in the [OpenWhisk](https://openwhisk.org) serverless project. Developers
in that project had a python script for doing similar comparisons, but
agreed we should extend it to a more general framework which could be
easily be extended for other lifecycle operation sequences, as the python
script was hardcoded to a specific set of operations.

## Usage
Using `bucketbench` to drive container operations against a specific
container runtime requires a configuration file written in a specific YAML
format.

The current driver implementations each support a small set
of lifecycle operations (defined as an interface in `driver/driver.go`), and
any benchmark definition can mix and match any of those operations within
reason. (Obviously operations must be ordered in a way supported by container
lifecycle--for example, you can't do `stop` prior to `run`.)

Specific command usage for the `bucketbench` program is as follows:
```
The YAML file provided via the --benchmark flag will determine which
lifecycle container commands to run against which container runtimes, specifying
iterations and number of concurrent threads. Results will be displayed afterwards.

Usage:
  bucketbench run [flags]

Flags:
  -b, --benchmark string   YAML file with benchmark definition
  -h, --help               help for run
  -s, --skip-limit         Skip 'limit' benchmark run
  -t, --trace              Enable per-container tracing during benchmark runs

Global Flags:
      --log-level string   set the logging level (info,warn,err,debug) (default "warn")
```

A common invocation for running the "basic" example benchmark might look like:

```
$ sudo ./bucketbench --log-level=debug run -b examples/basic.yaml
```

Let's look at the input YAML file format and define the components. Here's
the **basic.yaml** example:

```
name: Basic
image: alpine:latest
command: date
rootfs: /home/estesp/containers/alpine
detached: true
drivers:
  - 
   type: Docker
   threads: 5
   iterations: 15
  - 
   type: Runc
   threads: 5
   iterations: 50
commands:
  - run
  - stop
  - remove
```

The initial section sets up a name and a few key pieces of information required
for each engine to know **what** to run:
 - **name**: Give the benchmark a name. This will be used in output and logs.
 - **image**: Choose an image reference to be used by the image-based engine runtimes (containerd 1.0 and Docker). This can be any image reference accepted by the `docker pull` command. `bucketbench` will handle reconciling this reference to the format used by containerd 1.0 (e.g. `alpine` -> `docker.io/library/alpine:latest`)
 - **command**: *[Optional]* Specify an override for the image's default command that will be used for the image-based engine runtimes.
 - **rootfs**: For the `runc` and `ctr` (legacy containerd/0.2.x) drivers, you will need to provide an exploded rootfs and an OCI `config.json` since neither of those engines support image/registry interactions.
 - **detached**: Run the containers in detached/background mode.

The next two sections of the YAML provide 1) the configuration of which drivers
to execute the benchmark against, and 2) which lifecycle commands to run
against each engine.

#### Driver Configuration

Each driver has the following settings:
 - **type**: One of the four implemented drivers: `Runc`, `Docker`, `Containerd`, `Ctr`
 - **binary**: *[Optional]* Path to the binary (or in the case of containerd 1.0, UNIX socket path of the gRPC server) in case you want to use a custom binary. By default the standard binaries are used as found in the current `$PATH`
 - **threads**: Integer number of concurrent threads to run. The `bucketbench` method is to execute 1..n runs, where `n` is the number of threads and each run adds another concurrent thread. **Run 1** only has one thread and **Run N** will have `n` concurrent threads.
 - **iterations**: Number of containers to create in each thread and execute the listed commands against.

#### Command List

Finally, the YAML input needs to have a list of container lifecycle commands.
The following commands are accepted as input:

 - **run**: (aliases: **start**) create and start a container.
 - **pause**: pause a running container
 - **unpause**: (aliases: **resume**) resume a paused container
 - **stop**: (aliases: **kill**) stop/kill the running container processes
 - **remove**: (aliases: **erase**,**delete**) remove/delete a container instance

Note that `bucketbench` is not handling any formal state validation on the list
of commands. It is currently up to the user to provide a valid/sane ordered
list of container lifecycle commands. The container runtimes will error out on
incorrect command states (e.g. `stop` before `run`).

After the benchmark runs are complete, `bucketbench` currently provides basic
output to show the overall rate (iterations of the operations/second) for each
of the thread counts:

```
             Iter/Thd     1 thrd  2 thrds  3 thrds  4 thrds  5 thrds  6 thrds  7 thrds  8 thrds  9 thrds 10 thrds
Limit            1000    1171.24  1957.17  2101.13  2067.83  1827.92  1637.32  1257.57  1582.36  1306.08  1699.56
Basic:Docker       15       1.40     2.21     2.81
Basic:Runc         50       8.38    15.85    23.00
```

More detailed information is collected during the runs and a future PR to
`bucketbench` will provide the raw performance data in a consumable format for
end users.

To run `bucketbench` against `Runc`, `Containerd`, or the legacy `Ctr` driver
you must use `sudo` because of the requirements that those tools have for root
access. This tool does not manage the two daemon-based engines (containerd and
dockerd), and will fail if they are not up and running when the benchmark runs
begin.

The tool will start a significant number of containers against these daemons,
but attempts to fully cleanup after running each iteration.

## Development Notes

The `bucketbench` tool is most likely only valuable on amd64/linux, as
`containerd` and `runc` are delivered today as binaries for those platforms.
It will most likely build for other platforms, and if run against a tool like
Docker for Mac, would probably work against the Docker engine, but not
against `containerd` or `runc`.

All the necessary dependencies are vendored into the `bucketbench` tree, so
building should be as easy as `go build -o bucketbench .` Using `go install github.com/estesp/bucketbench`
should work as well.

## TODOs

 - UX/access to detailed statistics gathered (and currently unused) for each operation's metrics
 - Decide what to do with the `-trace` flag, which was only useful with a private build of `runc` which generated Go pprof traces. Possibly submit trace support to upstream runc.
