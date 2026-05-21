variable "prefix" {
  default = "tf-acc-test-file-storage-"
  type    = string
}

resource "upcloud_file_storage" "encrypted" {
  name              = "${var.prefix}-encrypted"
  size              = 250
  zone              = "fi-hel2"
  configured_status = "stopped"
  encrypt           = true
}
