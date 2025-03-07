task "install_nginx" {
    step "cmd" {
        command = "echo 'installing nginx'"
    }
}

task "configure_nginx" {
    step "cmd" {
        command = "echo 'configuring nginx'"
    }
}
