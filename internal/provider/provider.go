package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ tfsdk.Provider = &provider{}

// provider satisfies the tfsdk.Provider interface and usually is included
// with all Resource and DataSource implementations.
type provider struct {
	// client can contain the upstream provider SDK or HTTP client used to
	// communicate with the upstream service. Resource and DataSource
	// implementations can then make calls using this client.
	//
	// TODO: If appropriate, implement upstream provider SDK or HTTP client.
	// client vendorsdk.ExampleClient

	// configured is set to true at the end of the Configure method.
	// This can be used in Resource and DataSource implementations to verify
	// that the provider was previously configured.
	configured bool

	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

type providerData struct {
	FirstName   types.String    `tfsdk:"first_name"`
	LastName    types.String    `tfsdk:"last_name"`
	EmailAddr   types.String    `tfsdk:"email_address"`
	PhoneNumber types.String    `tfsdk:"phone_number"`
	CreditCard  *creditCardData `tfsdk:"credit_card"`
}

type creditCardData struct {
	CreditCardNumber types.Int64  `tfsdk:"number"`
	Cvv              types.Int64  `tfsdk:"cvv"`
	ExprDate         types.String `tfsdk:"date"`
	Zip              types.String `tfsdk:"zip"`
	CardType         types.String `tfsdk:"card_type"`
}

func (p *provider) Configure(ctx context.Context, req tfsdk.ConfigureProviderRequest, resp *tfsdk.ConfigureProviderResponse) {
	var data providerData
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// data.CreditCard.CardType.Value = "VISA"

	p.configured = true
}

func (p *provider) GetResources(ctx context.Context) (map[string]tfsdk.ResourceType, diag.Diagnostics) {
	return map[string]tfsdk.ResourceType{
		"dominos_order": resourceOrderType{},
	}, nil
}

func (p *provider) GetDataSources(ctx context.Context) (map[string]tfsdk.DataSourceType, diag.Diagnostics) {
	return map[string]tfsdk.DataSourceType{
		"dominos_address": dataSourceAddressType{},
		// "dominos_store":     dataSourceStoreType{},
		// "dominos_menu":      dataSourceMenuType{},
		// "dominos_menu_item": dataSourceMenuItemType{},
	}, nil
}

func (p *provider) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"email_address": {
				// MarkdownDescription: "Example provider attribute",
				Optional: true,
				Type:     types.StringType,
			},
			"first_name": {
				// MarkdownDescription: "Example provider attribute",
				Optional: true,
				Type:     types.StringType,
			},
			"last_name": {
				// MarkdownDescription: "Example provider attribute",
				Optional: true,
				Type:     types.StringType,
			},
			"phone_number": {
				// MarkdownDescription: "Example provider attribute",
				Optional: true,
				Type:     types.StringType,
			},
			"credit_card": {
				// MarkdownDescription: "Example provider attribute",
				Required:  true,
				Sensitive: true,
				Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
					"number": {
						Type:     types.Int64Type,
						Required: true,
					},
					"cvv": {
						Type:     types.Int64Type,
						Required: true,
					},
					"date": {
						Type:     types.StringType,
						Required: true,
					},
					"zip": {
						Type:     types.StringType,
						Required: true,
					},
					"card_type": {
						Type:     types.StringType,
						Optional: true,
					},
				}),
			},
		},
	}, nil
}

func New(version string) func() tfsdk.Provider {
	return func() tfsdk.Provider {
		return &provider{
			version: version,
		}
	}
}

// convertProviderType is a helper function for NewResource and NewDataSource
// implementations to associate the concrete provider type. Alternatively,
// this helper can be skipped and the provider type can be directly type
// asserted (e.g. provider: in.(*provider)), however using this can prevent
// potential panics.
func convertProviderType(in tfsdk.Provider) (provider, diag.Diagnostics) {
	var diags diag.Diagnostics

	p, ok := in.(*provider)

	if !ok {
		diags.AddError(
			"Unexpected Provider Instance Type",
			fmt.Sprintf("While creating the data source or resource, an unexpected provider type (%T) was received. This is always a bug in the provider code and should be reported to the provider developers.", p),
		)
		return provider{}, diags
	}

	if p == nil {
		diags.AddError(
			"Unexpected Provider Instance Type",
			"While creating the data source or resource, an unexpected empty provider instance was received. This is always a bug in the provider code and should be reported to the provider developers.",
		)
		return provider{}, diags
	}

	return *p, diags
}
