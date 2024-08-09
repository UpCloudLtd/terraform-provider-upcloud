resource "upcloud_storage_template" "template" {
  source_storage = "e0328f8a-9944-406b-99c3-656dcc03e671"
  title          = "custom-storage-template"

  labels = {
    os    = "linux"
    usage = "example"
  }
}
