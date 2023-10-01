resource "null_resource" "first" {
  provisioner "local-exec" {
    command = "echo first"
  }
}
