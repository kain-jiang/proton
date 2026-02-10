target "ctx-static-web" {
    context = "deploy-web/deploy-web-static"
    tags = [
        "ctx-static-web"
    ]
}



target "ctx-component-manage-src" {
    dockerfile-inline = "FROM scratch\nCOPY ./component-manage /\n"
    tags = [
        "ctx-component-manage-src"
    ]
}

group "default" {
    targets = [
        "ctx-static-web",
        "ctx-component-manage-src"
    ]
}