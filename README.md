# voki

Voki configuration management

# Usage

Change the host in `target.example.hcl` to a SSH reachable address and then run

```sh
./voki run target.example.hcl
```

## Specification

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