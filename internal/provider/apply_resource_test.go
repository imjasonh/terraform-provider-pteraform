// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccExampleResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: `
resource "pteraform_apply" "first" {
	working_dir = "testdata/first"
}

resource "pteraform_apply" "second" {
	working_dir = "testdata/second"
	args = ["-var=value=cool"]
}
`,
			Check: resource.ComposeAggregateTestCheckFunc(
			// TODO: check stuff?
			),
		}},
	})

	for _, fn := range []string{"first", "second"} {
		if _, err := os.Stat(filepath.Join("testdata", fn, "terraform.tfstate")); err != nil {
			t.Errorf("expected terraform.tfstate to exist in %s, got error: %s", fn, err)
		}
	}
}
