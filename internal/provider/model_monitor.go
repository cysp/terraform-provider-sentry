package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/jianyuan/terraform-provider-sentry/internal/apiclient"
	"github.com/jianyuan/terraform-provider-sentry/internal/tfutils"
)

type MonitorResourceModel struct {
	Id           types.String `tfsdk:"id"`
	Organization types.String `tfsdk:"organization"`
	Type         types.String `tfsdk:"type"`
	Name         types.String `tfsdk:"name"`
	Owner        types.String `tfsdk:"owner"`
	Project      types.String `tfsdk:"project"`
	Slug         types.String `tfsdk:"slug"`
	Config       types.Object `tfsdk:"config"`
	IsMuted      types.Bool   `tfsdk:"is_muted"`
	Status       types.String `tfsdk:"status"`
}

func (m MonitorResourceModel) Attributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id":           ResourceIdAttribute(),
		"organization": ResourceOrganizationAttribute(),
		"project":      ResourceProjectAttribute(),
		"type": schema.StringAttribute{
			Required: true,
		},
		"name": schema.StringAttribute{
			Required: true,
		},
		"owner": schema.StringAttribute{
			Optional: true,
		},
		"slug": schema.StringAttribute{
			Required: true,
		},
		"config": MonitorConfigResourceModel{}.Schema(true),
		"is_muted": schema.BoolAttribute{
			Optional: true,
		},
		"status": schema.StringAttribute{
			Optional: true,
			Computed: true,
		},
	}
}

func (m *MonitorResourceModel) Fill(ctx context.Context, organization string, monitor apiclient.Monitor) (diags diag.Diagnostics) {
	path := path.Empty()

	m.Id = types.StringValue(monitor.Id)
	m.Type = types.StringValue(string(monitor.Type))
	m.Name = types.StringValue(monitor.Name)
	m.Slug = types.StringValue(monitor.Slug)
	m.Organization = types.StringValue(organization)
	m.Project = types.StringValue(monitor.Project.Slug)
	m.Status = types.StringValue(monitor.Status)
	m.IsMuted = types.BoolValue(monitor.IsMuted)

	if monitor.Owner == nil {
		m.Owner = types.StringNull()
	} else {
		m.Owner = types.StringValue(string(monitor.Owner.Type) + ":" + monitor.Owner.Id)
	}

	var config MonitorConfigResourceModel
	m.Config.As(ctx, &config, basetypes.ObjectAsOptions{})
	diags.Append(config.Fill(ctx, path.AtName("config"), monitor.Config)...)
	m.Config = tfutils.MergeDiagnostics(types.ObjectValueFrom(ctx, config.AttributeTypes(), config))(&diags)

	return
}

type MonitorConfigResourceModel struct {
	ScheduleCrontab       types.String `tfsdk:"schedule_crontab"`
	ScheduleInterval      types.Object `tfsdk:"schedule_interval"`
	CheckinMargin         types.Int64  `tfsdk:"checkin_margin"`
	MaxRuntime            types.Int64  `tfsdk:"max_runtime"`
	Timezone              types.String `tfsdk:"timezone"`
	FailureIssueThreshold types.Int64  `tfsdk:"failure_issue_threshold"`
	RecoveryThreshold     types.Int64  `tfsdk:"recovery_threshold"`
	AlertRuleId           types.Int64  `tfsdk:"alert_rule_id"`
}

func (m MonitorConfigResourceModel) Schema(required bool) schema.Attribute {
	return schema.SingleNestedAttribute{
		Required:   required,
		Attributes: m.SchemaAttributes(),
	}
}

func (m MonitorConfigResourceModel) SchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"schedule_crontab": schema.StringAttribute{
			Optional: true,
			Validators: []validator.String{
				stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("schedule_interval")),
			},
		},
		"schedule_interval": schema.SingleNestedAttribute{
			Optional:   true,
			Attributes: MonitorConfigIntervalScheduleResourceModel{}.SchemaAttributes(),
			Validators: []validator.Object{
				objectvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("schedule_crontab")),
			},
		},
		"checkin_margin": schema.Int64Attribute{
			Optional: true,
		},
		"max_runtime": schema.Int64Attribute{
			Optional: true,
		},
		"timezone": schema.StringAttribute{
			Optional: true,
		},
		"failure_issue_threshold": schema.Int64Attribute{
			Optional: true,
		},
		"recovery_threshold": schema.Int64Attribute{
			Optional: true,
		},
		"alert_rule_id": schema.Int64Attribute{
			Optional: true,
		},
	}
}

func (m *MonitorConfigResourceModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"schedule_crontab":        types.StringType,
		"schedule_interval":       types.ObjectType{AttrTypes: (&MonitorConfigIntervalScheduleResourceModel{}).AttributeTypes()},
		"checkin_margin":          types.Int64Type,
		"max_runtime":             types.Int64Type,
		"timezone":                types.StringType,
		"failure_issue_threshold": types.Int64Type,
		"recovery_threshold":      types.Int64Type,
		"alert_rule_id":           types.Int64Type,
	}
}

type MonitorConfigIntervalScheduleResourceModel struct {
	Year   types.Int64 `tfsdk:"year"`
	Month  types.Int64 `tfsdk:"month"`
	Week   types.Int64 `tfsdk:"week"`
	Day    types.Int64 `tfsdk:"day"`
	Hour   types.Int64 `tfsdk:"hour"`
	Minute types.Int64 `tfsdk:"minute"`
}

func (m MonitorConfigIntervalScheduleResourceModel) SchemaAttributes() map[string]schema.Attribute {
	attributeNames := []string{"year", "month", "week", "day", "hour", "minute"}

	attributes := make(map[string]schema.Attribute, len(attributeNames))

	for _, name := range attributeNames {
		var conflictingPaths []path.Expression

		for _, conflictingName := range attributeNames {
			if conflictingName != name {
				conflictingPaths = append(conflictingPaths, path.MatchRelative().AtParent().AtName(conflictingName))
			}
		}

		attributes[name] = schema.Int64Attribute{
			Optional: true,
			Validators: []validator.Int64{
				int64validator.AtLeast(1),
				int64validator.ConflictsWith(conflictingPaths...),
			},
		}
	}

	return attributes
}

func (m *MonitorConfigIntervalScheduleResourceModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"year":   types.Int64Type,
		"month":  types.Int64Type,
		"week":   types.Int64Type,
		"day":    types.Int64Type,
		"hour":   types.Int64Type,
		"minute": types.Int64Type,
	}
}

func (m *MonitorConfigResourceModel) Fill(ctx context.Context, path path.Path, config apiclient.MonitorConfig) (diags diag.Diagnostics) {
	switch config.ScheduleType {
	case apiclient.Crontab:
		schedule, scheduleErr := config.Schedule.AsMonitorConfigScheduleString()
		if scheduleErr != nil {
			diags.AddAttributeError(path.AtName("schedule"), "Invalid schedule", scheduleErr.Error())
			break
		}
		m.ScheduleCrontab = types.StringValue(schedule)
		m.ScheduleInterval = types.ObjectNull((&MonitorConfigIntervalScheduleResourceModel{}).AttributeTypes())
	case apiclient.Interval:
		schedule, scheduleErr := config.Schedule.AsMonitorConfigScheduleInterval()
		if scheduleErr != nil {
			diags.AddAttributeError(path.AtName("schedule"), "Invalid schedule", scheduleErr.Error())
			break
		}
		parsedSchedule := tfutils.MergeDiagnostics(parseMonitorConfigIntervalSchedule(schedule))(&diags)
		m.ScheduleCrontab = types.StringNull()
		m.ScheduleInterval = tfutils.MergeDiagnostics(types.ObjectValueFrom(ctx, (&MonitorConfigIntervalScheduleResourceModel{}).AttributeTypes(), parsedSchedule))(&diags)
	default:
		diags.AddAttributeError(path.AtName("schedule"), "Invalid schedule type", string(config.ScheduleType))
	}

	m.CheckinMargin = types.Int64Value(config.CheckinMargin)
	m.MaxRuntime = types.Int64PointerValue(config.MaxRuntime)
	m.Timezone = types.StringPointerValue(config.Timezone)
	m.FailureIssueThreshold = types.Int64PointerValue(config.FailureIssueThreshold)
	m.RecoveryThreshold = types.Int64PointerValue(config.RecoveryThreshold)

	if config.AlertRuleId != nil {
		alertRuleId, alertRuleIdErr := config.AlertRuleId.Int64()
		if alertRuleIdErr != nil {
			diags.AddAttributeError(path.AtName("alert_rule_id"), "Invalid alert rule ID", alertRuleIdErr.Error())
		}
		m.AlertRuleId = types.Int64Value(alertRuleId)
	}

	return
}

func parseMonitorConfigIntervalSchedule(m apiclient.MonitorConfigScheduleInterval) (MonitorConfigIntervalScheduleResourceModel, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	rm := MonitorConfigIntervalScheduleResourceModel{}

	if len(m) != 2 {
		diags.AddError("Invalid schedule", "Invalid schedule")
		return rm, diags
	}

	var number int64
	number, ok := m[0].(int64)
	if !ok {
		diags.AddError("Invalid schedule", "Invalid schedule")
		return rm, diags
	}

	var unit string
	unit, ok = m[1].(string)
	if !ok {
		diags.AddError("Invalid schedule", "Invalid schedule")
		return rm, diags
	}

	switch unit {
	case "year":
		rm.Year = types.Int64Value(number)
	case "month":
		rm.Month = types.Int64Value(number)
	case "week":
		rm.Week = types.Int64Value(number)
	case "day":
		rm.Day = types.Int64Value(number)
	case "hour":
		rm.Hour = types.Int64Value(number)
	case "minute":
		rm.Minute = types.Int64Value(number)
	default:
		diags.AddError("Invalid schedule", "Invalid schedule")
	}

	return rm, diags
}
