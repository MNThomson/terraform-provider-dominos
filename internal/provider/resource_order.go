package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ provider.ResourceType = resourceOrderType{}
var _ resource.Resource = resourceOrder{}
var _ resource.ResourceWithImportState = resourceOrder{}

type resourceOrderType struct{}

func (t resourceOrderType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		// This description is used by the documentation generator and the language server.
		Description: `
This is it! This will order you your pizzas!

As far as I know there is no way to cancel a dominos order programmatically, so if you made a mistake, you'll have to call the store.
You should receive an email confirmation almost instantly, and that email will have the store's phone number in it.
		`,
		Attributes: map[string]tfsdk.Attribute{
			"api_object": {
				Description: "The computed json payload for the specified address.",
				Required:    true,
				Type:        types.StringType,
			},
			"item_codes": {
				Description: "An array of menu items to order.",
				Required:    true,
				Type: types.ListType{
					ElemType: types.StringType,
				},
			},
			"store_id": {
				Description: "The ID of the store that the order is for.",
				Required:    true,
				Type:        types.Int64Type,
			},
			"price_only": {
				Description: "DRY RUN: This will only display the total price of the order (and not actually order).",
				Optional:    true,
				Type:        types.BoolType,
			},
			"total_price": {
				Description: "The computed total price of the order.",
				Computed:    true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					resource.UseStateForUnknown()},
				Type: types.NumberType,
			},
		},
	}, nil
}

func (t resourceOrderType) NewResource(ctx context.Context, in provider.Provider) (resource.Resource, diag.Diagnostics) {
	provider, diags := convertProviderType(in)

	return resourceOrder{
		provider: provider,
	}, diags
}

type resourceOrderData struct {
	AddressAPIObj types.String `tfsdk:"api_object"`
	ItemCodes     types.List   `tfsdk:"item_codes"`
	StoreID       types.Int64  `tfsdk:"store_id"`
	PriceOnly     types.Bool   `tfsdk:"price_only"`
	TotalPrice    types.Number `tfsdk:"total_price"`
}

type resourceOrder struct {
	provider dominosProvider
}

func (r resourceOrder) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data resourceOrderData

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r resourceOrder) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data resourceOrderData

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r resourceOrder) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data resourceOrderData

	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r resourceOrder) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data resourceOrderData

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
}

func (r resourceOrder) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
