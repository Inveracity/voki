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

    Or via `voki-config.hcl` see the [configuration section](#configuration)

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
- `data` a string to write to the destination file. If `source` is set, then `data` is ignored.
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

it's also possible to supply data, which enables use of the inline functions `file()` and `template()`

```hcl
target "mytarget" {
    user = "root"
    host = "address:22"

    step "file" {
      data = "content for the file"
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

### Defining variables

The target file can have variables set in the following way

```hcl
// target.hcl
name = "my name"

target "myserver" {
    user = "root"
    host = "xyz:22"

    step "cmd" {
        command = "echo ${name}"
    }
}
```

## Configuration

Configuration and environment variables can be specified via a `voki-config.hcl` file.

```sh
# voki-config.hcl
user="myuser"
vault-address="http://127.0.0.1:8200"
vault-token="123456"
```

or with `VOKI_` prefixed variables

```sh
VOKI_USER="me" voki run target.hcl
```

With this set, the target section no longer requires the `user` variables.

## Multiple targets (and parallelization)

Voki supports specifying multiple targets and run them sequentially or in parallel.

Default is sequential by simply invoking multiple target files:

```sh
voki run target1.hcl target2.hcl ... etc.
```

and to run them in parallel add the `-p <number>` specifying how many to run in parallel:

```sh
voki run -p 2 target1.hcl target2.hcl ... etc.
```

## Vault integration

Install vault and run it in _dev mode_ to test locally.

```sh
cd /tmp
curl -L https://releases.hashicorp.com/vault/1.19.0/vault_1.19.0_linux_amd64.zip -O
unzip vault_1.19.0_linux_amd64.zip
rm -f LICENSE.txt
rm -f vault_1.19.0_linux_amd64.zip
install vault ~/.local/bin/vault
rm -f vault
cd

export VAULT_DEV_ROOT_TOKEN_ID=123456
vault server -dev
export VAULT_ADDR='http://127.0.0.1:8200'
```

add a secret to vault

```sh
vault kv put -mount=secret voki hello=world
```

Now use the secret in a target file.

Notice the path in the Vault block, it's required to access the secret i.e. `vault.voki.hello`. If the path was "jazz" it would be `vault.jazz.hello`.

```hcl
vault {
    mountpath = "secret"
    path = "voki"
}

target "myserver" {
    host = "127.0.0.1:22"
    user = "root"

    step "cmd" {
        command = "echo ${vault.voki.hello}"
    }
}
```

and invoke voki

```sh
VOKI_VAULT_TOKEN=123456 VOKI_VAULT_ADDR="http://127.0.0.1:8200" voki run target.hcl
```

It's also possible to grab secrets from different paths

add another secret in a different path

```sh
vault kv put -mount=secret another hello=world
```

```hcl
vault {
    mountpath = "secret"
    path = "voki"
}

vault {
    mountpath = "secret"
    path = "another"
}

target "myserver" {
    host = "127.0.0.1:22"
    user = "root"

    step "cmd" {
        command = "echo ${vault.voki.hello}"
    }

    step "cmd" {
        command = "echo ${vault.another.hello}"
    }
}
```

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
