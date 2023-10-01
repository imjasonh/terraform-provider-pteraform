# `terraform-provider-pteraform`

This Terraform provider allows you to invoke `terraform apply` from within a Terraform resource.

Sometimes I have good ideas, and sometimes I have bad ideas, and sometimes I can't tell which is which.

## Usage

```hcl
provider "pteraform" {}

resource "pteraform_apply" "first" {
  working_dir = "${path.module}/first"
}

resource "pteraform_apply" "second" {
  working_dir = "${path.module}/second"
}
```

Applying this config will run `terraform apply -auto-approve` in both directories.

If you change the contents of one of those directories, only that directory will be applied again.
`terraform plan` will show you which directories changed (not very usefully).

Do not use this in anything approaching real life.

### Pteraform...?

`terraform` is a reserved provider name. I know right?
