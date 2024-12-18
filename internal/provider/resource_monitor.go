package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jianyuan/go-sentry/v2/sentry"
)

var _ resource.Resource = &MonitorResource{}
var _ resource.ResourceWithConfigure = &MonitorResource{}
var _ resource.ResourceWithImportState = &MonitorResource{}

func NewMonitorResource() resource.Resource {
	return &MonitorResource{}
}

type MonitorResource struct {
	baseResource
}

type MonitorResourceModel struct {
	Id           types.String `tfsdk:"id"`
	Organization types.String `tfsdk:"organization"`
	// Project      types.String `tfsdk:"project"`
	Name types.String `tfsdk:"name"`
	// Type    types.String `tfsdk:"type"`
	Slug    types.String `tfsdk:"slug"`
	Status  types.String `tfsdk:"status"`
	Owner   types.String `tfsdk:"owner"`
	IsMuted types.Bool   `tfsdk:"is_muted"`
}

func (m *MonitorResourceModel) Fill(organization string, monitor sentry.Monitor) error {
	m.Id = types.StringValue(monitor.ID)
	m.Organization = types.StringValue(organization)
	// m.Project = types.StringValue(project)
	m.Name = types.StringValue(monitor.Name)
	// m.Type = types.StringValue(monitor.Type)
	m.Slug = types.StringValue(monitor.Slug)
	m.Status = types.StringValue(monitor.Status)
	m.Owner = types.StringValue(monitor.Owner.Type + ":" + monitor.Owner.ID)
	m.IsMuted = types.BoolValue(monitor.IsMuted)

	return nil
}

func (r *MonitorResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_monitor"
}

func (r *MonitorResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Return a client monitor bound to a project.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of this resource.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"organization": schema.StringAttribute{
				MarkdownDescription: "The slug of the organization the resource belongs to.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the monitor.",
				Required:            true,
			},
			// "type": schema.StringAttribute{
			// 	MarkdownDescription: "The type of the monitor.",
			// 	Required:            true,
			// },
			"slug": schema.StringAttribute{
				MarkdownDescription: "The slug of the monitor.",
				Optional:            true,
				Computed:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "The status of the monitor.",
				Optional:            true,
				Computed:            true,
			},
			"owner": schema.StringAttribute{
				MarkdownDescription: "The owner of the monitor.",
				Optional:            true,
			},
			"is_muted": schema.BoolAttribute{
				MarkdownDescription: "The mute status of the monitor.",
				Optional:            true,
			},
		},
	}
}

func (r *MonitorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data MonitorResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := &sentry.CreateOrUpdateMonitorParams{
		Type:    "cron_job",
		Name:    data.Name.ValueString(),
		Slug:    data.Slug.ValueStringPointer(),
		Status:  data.Status.ValueStringPointer(),
		Owner:   data.Owner.ValueStringPointer(),
		IsMuted: data.IsMuted.ValueBoolPointer(),
	}

	monitor, _, err := r.client.Monitors.Create(
		ctx,
		data.Organization.ValueString(),
		params,
	)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Create error: %s", err.Error()))
		return
	}
	if err := data.Fill(data.Organization.ValueString(), *monitor); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Fill error: %s", err.Error()))
		return
	}

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MonitorResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data MonitorResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	monitor, apiResp, err := r.client.Monitors.Get(
		ctx,
		data.Organization.ValueString(),
		data.Id.ValueString(),
	)
	if apiResp.StatusCode == http.StatusNotFound {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Not found: %s", err.Error()))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Read error: %s", err.Error()))
		return
	}

	if err := data.Fill(data.Organization.ValueString(), *monitor); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Fill error: %s", err.Error()))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MonitorResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data MonitorResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := &sentry.CreateOrUpdateMonitorParams{
		Type:    "cron_job",
		Name:    data.Name.ValueString(),
		Slug:    data.Slug.ValueStringPointer(),
		Status:  data.Status.ValueStringPointer(),
		Owner:   data.Owner.ValueStringPointer(),
		IsMuted: data.IsMuted.ValueBoolPointer(),
	}
	monitor, apiResp, err := r.client.Monitors.Update(
		ctx,
		data.Organization.ValueString(),
		data.Id.ValueString(),
		params,
	)
	if apiResp.StatusCode == http.StatusNotFound {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Not found: %s", err.Error()))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Update error: %s", err.Error()))
		return
	}

	if err := data.Fill(data.Organization.ValueString(), *monitor); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Fill error: %s", err.Error()))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MonitorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data MonitorResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiResp, err := r.client.Monitors.Delete(
		ctx,
		data.Organization.ValueString(),
		data.Id.ValueString(),
	)
	if apiResp.StatusCode == http.StatusNotFound {
		return
	}

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Delete error: %s", err.Error()))
		return
	}
}

func (r *MonitorResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	organization, id, err := splitTwoPartID(req.ID, "organization", "monitor-id")
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", fmt.Sprintf("Error parsing ID: %s", err.Error()))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(
		ctx, path.Root("organization"), organization,
	)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(
		ctx, path.Root("id"), id,
	)...)
}
