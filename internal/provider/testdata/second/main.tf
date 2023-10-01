variable "value" {
  type    = string
  default = "second"
}

resource "null_resource" "second" {
  provisioner "local-exec" {
    command = "echo ${var.value}"
  }
}
