targets {
    user = "root" // SSH username to connect to the remote host with
    name = "myserver" // Name to show in the terminal
    host = "127.0.0.1:22" // Remote host to connect to
    cmd = "echo 'Hello, World!'" // Command to run on the remote host
}

targets {
    user = "root"
    name = "local"
    host = "localhost:22"
}
