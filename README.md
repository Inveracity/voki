# voki

Voki configuration management.

Voki uses SSH under the hood to connect to remote machines.

> [!NOTE]
> As of right now Voki requires an active SSH Agent to connect to machines.

## Usage

Change the host in `examples/target.hcl` to a SSH reachable address and then run

```sh
./voki run examples/target.hcl
```

## Example

```hcl
target "myserver" {
    user = "root" // SSH username to connect to the remote host with
    host = "ADDRESS:22" // Remote host to connect to

    step "cmd" {
        command = "echo 'hello world'"
    }

    step "cmd" {
        command = file("hello.sh") // requires a hello.sh script to exist
    }
}
```

## Specifications

Voki configuration file consist of Targets and Task files.

The target configurations have the following structure:

`Target "label" -> Step "action"`

and Task configurations only includes:

`Step "action"`

Targets and tasks files can have any name desired.

### Targets

_Target_ takes the following variables to configure the SSH connection.

```hcl
target "a label for the configuration" {
    user = "root"
    host = "address:22"

    // steps go here
}
```

- `user` the username to connect to the remote host.

    The user variable can also be provided via a flag `voki -u USERNAME`

    Or via `.voki.env` see the [configuration section](#configuration)

- `host` the address including the port number separated by a colon: `ipaddress:port`

### Steps

Steps has a label for which _Action_ to perform for the remote host.

```hcl
target "mytarget" {
    user = "root"
    host = "address:22"

    step "cmd" {
        command = "hello world"
    }
}
```

In the example above the _Step_ takes the label `"cmd"` which tells voki what the _Action_ is.

See the actions available below.

### Actions

Actions are taken based on the label applied to the step as shown in the _Steps_ section above.

#### Action: Cmd

Cmd runs a command or multiple commands on the configured target.

Cmd takes the following variables:

- `command` commands to run on the remote host.
- `sudo` boolean value to execute commands with sudo (defaults to `false`)
- `shell` which shell to use, i.e. `bash` or `sh` (defaults to `bash`)

```hcl
target "mytarget" {
    user = "root"
    host = "address:22"

    step "cmd" {
        command = "echo 'hello world'"
    }

    // Multiline input for multiple commands and scripts.
    step "cmd" {
        sudo = true
        shell = "sh"
        command = <<-EOT
            echo "1"
            echo "2"
        EOT
    }
}
```

See the _Inline functions_ for running scripts with the Cmd action.

#### Action: File

File copies a file from the local filesystem to the remote filesystem.

File takes the following variables:

- `source` the path on the host voki is executed on.
- `destination` the path on the remote host where the file should be copied to.
- `mode` the permissions on the file on the remote host.

```hcl
target "mytarget" {
    user = "root"
    host = "address:22"

    step "file" {
      source = "myfile.conf"
      destination = "/var/myfile.conf"
      mode = "0644"
    }
}
```

#### Action: Task

Task is a special action that reads a separate configuration file with steps.

In the `target.hcl` file define the `step "task" {}` an use the inline function `file()` to read the `task.hcl` file.

- `target.hcl`

    ```hcl
    target "mytarget" {
        user = "root"
        host = "address:22"

        step "task" {
            task = file("task.hcl")
        }
    }
    ```

- `task.hcl`

    ```hcl
    step "cmd" {
        command = "echo 'hello task'"
    }
    ```

A task file does not have a _Target_ specification, only _Steps_.

Task files can also include nested task actions if so desired.

### Inline Functions

#### Function: file()

Load a text file such as a script and pass it without modification.

In the below example a script is loaded and executed on a target.

```hcl
// target.hcl
target "myserver" {
    user = "root"
    host = "xyz:22"

    step "cmd" {
        command = file("hello.sh")
    }
}
```

and the contents of `hello.sh`

```sh
echo "hello world"
```

#### Function: template()

Load a text file such as a script and pass dynamic data to be rendered before use.

The template rendering uses Go's builtin [html/template](https://pkg.go.dev/html/template).

In the below example a script has a key/value pair passed in that will be rendered before executed on a target.

```hcl
// target.hcl
target "myserver" {
    user = "root"
    host = "xyz:22"

    step "cmd" {
        command = template("hello.sh.tpl", {
            Name: "world!"
        })
    }
}
```

in the `hello.sh.tpl` the `Name` is being passed in before execution.

```sh
echo "hello {{ .Name }}"
```

## Configuration

Some variables can be specified via a `.voki.env` file.

```sh
# .voki.env
user="myuser"
```

With this set, the target section no longer requires the `user` variables.

## Build from source

Ensure Go 1.23 or newer is installed.

```sh
make build

./bin/voki --help
```

Install into `/home/$(USER)/.local/bin`

```sh
make install
```
