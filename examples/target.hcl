import {
  file = "examples/task.hcl"
}

import {
  file = "examples/task2.hcl"
}

target "myserver" {
    user = "root" // SSH username to connect to the remote host with
    host = "127.0.0.1:22" // Change this to the IP address of your remote host

    // Simple command
    step "cmd" {
        command = "echo 'Hello step 1'"
    }

    // Multiline command
    step "cmd" {
        command = <<-EOT
        echo "hello"
        echo "step 2"
        EOT
    }

    // Script file
    step "cmd" {
        command = file("examples/hello.sh")
    }


    // Run imported steps
    apply {
      use = ["install_nginx", "configure_nginx", "test"]
    }
}
