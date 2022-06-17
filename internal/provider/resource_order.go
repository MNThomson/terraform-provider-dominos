package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ tfsdk.ResourceType = resourceOrderType{}
var _ tfsdk.Resource = resourceOrder{}
var _ tfsdk.ResourceWithImportState = resourceOrder{}

type resourceOrderType struct{}

func (t resourceOrderType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Example resource",
		Attributes: map[string]tfsdk.Attribute{
			"address_api_object": {
				Required: true,
				Type:     types.StringType,
			},
			"item_codes": {
				Required: true,
				Type: types.ListType{
					ElemType: types.StringType,
				},
			},
			"store_id": {
				Required: true,
				Type:     types.Int64Type,
			},
			"price_only": {
				Optional: true,
				Type:     types.BoolType,
			},
			"total_price": {
				Computed: true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.UseStateForUnknown(),
				},
				Type: types.NumberType,
			},
		},
	}, nil
}

func (t resourceOrderType) NewResource(ctx context.Context, in tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	provider, diags := convertProviderType(in)

	return resourceOrder{
		provider: provider,
	}, diags
}

type resourceOrderData struct {
	AddressAPIObj types.String `tfsdk:"address_api_object"`
	ItemCodes     types.List   `tfsdk:"item_codes"`
	StoreID       types.Int64  `tfsdk:"store_id"`
	PriceOnly     types.Bool   `tfsdk:"price_only"`
	TotalPrice    types.Number `tfsdk:"total_price"`
}

type resourceOrder struct {
	provider provider
}

func (r resourceOrder) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	var data resourceOrderData

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r resourceOrder) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	var data resourceOrderData

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r resourceOrder) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	var data resourceOrderData

	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r resourceOrder) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var data resourceOrderData

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
}

func (r resourceOrder) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	tfsdk.ResourceImportStatePassthroughID(ctx, tftypes.NewAttributePath().WithAttributeName("id"), req, resp)
}
