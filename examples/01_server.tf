provider "upcloud" {
    username = "foo"
    password = "bar"
}

resource "upcloud_server" "test" {

    # System hostname
    hostname = "my-awesome-hostname"

    # Target datacenter
    zone = "fi-hel1"

    # Template name or UUID
    template = "CentOS 7.0"

    # Number of vCPUs
    cpu = 2

    # Amount of memory in MB
    mem = 4096

    # OS root disk size
    os_disk_size = 20

}
