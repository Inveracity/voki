---
layout: page
title: "Local Mode"
permalink: /localmode
---

Local mode is where the target is the same host as where voki is executed.

### Local mode configuration

To run in local mode simply set the target host value to `"localhost"`

```hcl
target "here" {
    host = "localhost"
    user = "root"

    step "cmd" {
        command = "echo world"
    }
}
```

### Local environment variables

In local mode it can be useful to reference environment variables on the system.

This can be done with `${env.THE_VARIABLE}`.

Notice the `user` variable can be disregarded as it will simply run as the current user.

```hcl
target "here" {
    host = "localhost"

    step "cmd" {
        // sudo = true
        command = "whoami && echo ${env.USER}"
    }
}
```
