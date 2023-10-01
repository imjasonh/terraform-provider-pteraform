// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &ApplyResource{}
var _ resource.ResourceWithImportState = &ApplyResource{}

func NewApplyResource() resource.Resource {
	return &ApplyResource{}
}

// ApplyResource defines the resource implementation.
type ApplyResource struct{}

// ApplyResourceModel describes the resource data model.
type ApplyResourceModel struct {
	WorkingDir types.String `tfsdk:"working_dir"`
	Args       types.List   `tfsdk:"args"`
	Id         types.String `tfsdk:"id"`
}

func (m *ApplyResourceModel) ID() (string, error) {
	f, err := os.Open(filepath.Join(m.WorkingDir.ValueString(), "terraform.tfstate"))
	if err != nil {
		return "", fmt.Errorf("Unable to open terraform.tfstate, got error: %s", err)
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("Unable to read terraform.tfstate, got error: %s", err)
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func (r *ApplyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_apply"
}

func (r *ApplyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Terraform Apply resource",

		Attributes: map[string]schema.Attribute{
			"working_dir": schema.StringAttribute{
				MarkdownDescription: "What directory to run `terraform apply` in.",
				Required:            true,
			},
			"args": schema.ListAttribute{
				MarkdownDescription: "Arguments to pass to `terraform apply`.",
				ElementType:         basetypes.StringType{},
				Optional:            true,
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Identifier of the resource.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func (r *ApplyResource) Configure(context.Context, resource.ConfigureRequest, *resource.ConfigureResponse) {
}

func (r *ApplyResource) doApply(ctx context.Context, data ApplyResourceModel) error {
	var buf bytes.Buffer

	// terraform init
	{
		cmd := exec.CommandContext(ctx, "terraform", "init")
		cmd.Dir = data.WorkingDir.ValueString()
		cmd.Stdout = &buf
		cmd.Stderr = &buf
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("terraform init failed, got error: %s, output: %s", err, buf.String())
		}
		buf.Reset()
	}

	// terraform apply -auto-approve
	{
		var args []string
		if diag := data.Args.ElementsAs(ctx, &args, false); diag.HasError() {
			return fmt.Errorf("errors getting args: %v", diag.Errors())
		}
		cmd := exec.CommandContext(ctx, "terraform", append([]string{"apply", "-auto-approve"}, args...)...)
		cmd.Dir = data.WorkingDir.ValueString()
		cmd.Stdout = &buf
		cmd.Stderr = &buf
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("terraform apply failed, got error: %s, output: %s", err, buf.String())
		}
		buf.Reset()
	}
	return nil
}

func (r *ApplyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ApplyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.doApply(ctx, data); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to run terraform apply, got error: %s", err))
	}

	id, err := data.ID()
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get ID, got error: %s", err))
	}
	data.Id = basetypes.NewStringValue(id)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ApplyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ApplyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := data.ID()
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get ID, got error: %s", err))
	}
	data.Id = basetypes.NewStringValue(id)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ApplyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ApplyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.doApply(ctx, data); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to run terraform apply, got error: %s", err))
	}

	id, err := data.ID()
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get ID, got error: %s", err))
	}
	data.Id = basetypes.NewStringValue(id)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ApplyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ApplyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Nothing to delete. Run `terraform destroy`? 🤷‍♂️
}

func (r *ApplyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
