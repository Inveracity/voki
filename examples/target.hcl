import {
  file = "examples/task.hcl"
}

target "myserver" {
    user = "root" // SSH username to connect to the remote host with
    host = "127.0.0.1:22" // Change this to the IP address of your remote host

    step "cmd" {
        command = "echo 'Hello step 1'"
    }

    step "cmd" {
        command = <<-EOT
        echo "hello"
        echo "step 2"
        EOT
    }


    apply {
      use = ["install_nginx", "configure_nginx"]
    }
}
