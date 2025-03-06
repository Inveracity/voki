target "myserver" {
    user = "root" // SSH username to connect to the remote host with
    host = "127.0.0.1:22" // Remote host to connect to

    step "cmd" {
        command = "echo 'Hello step 1'" 
    }

    step "cmd" {
        command = <<-EOT
        echo "hello"
        echo "step 2"
        EOT
    }
}

