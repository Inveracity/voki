step "cmd" {
    command = template("examples/hello.sh.tpl", {
        Name: "WORLD!"
    })
}
