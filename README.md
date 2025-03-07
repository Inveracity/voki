# voki

Voki configuration management

# Usage

Change the host in `target.example.hcl` to a SSH reachable address and then run

```sh
./voki run examples/target.hcl
```

## Specification example

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

## Inline functions

### file()

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

### template()

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
