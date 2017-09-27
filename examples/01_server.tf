provider "upcloud" {
    username = "foo"
    password = "bar"
}

resource "upcloud_server" "test" {
    hostname = "my-awesome-hostname"
    title = "my awesome title"
    zone = "fi-hel1"
}
