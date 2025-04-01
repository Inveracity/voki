address = "127.0.0.1:22"

target "myserver" {
  user = "root"         // SSH username to connect to the remote host with
  host = var.address // Change this to the IP address of your remote host

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

  // Template file
  step "cmd" {
    command = template("examples/hello.sh.tpl", {
      Name : "WORLD!"
    })
  }

  // Import steps from another file
  step "task" {
    task = file("task.hcl")
  }

  // Copy file to remote host
  step "file" {
    source      = "myfile.sh.tpl"
    destination = "/tmp/myfile.sh"
    mode        = "0755"
  }
}
