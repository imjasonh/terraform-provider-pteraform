resource "pteraform_apply" "first" {
  working_dir = "${path.module}/first"
}

resource "pteraform_apply" "second" {
  working_dir = "${path.module}/second"
}
