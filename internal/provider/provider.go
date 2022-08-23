package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ provider.Provider = &dominosProvider{}

// dominosProvider satisfies the provider.Provider interface and usually is included
// with all Resource and DataSource implementations.
type dominosProvider struct {
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

	providerdata providerData
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
	PostalCode       types.String `tfsdk:"postal_code"`
	CardType         types.String `tfsdk:"card_type"`
}

func (p *dominosProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data providerData
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	data.CreditCard.CardType = types.String{Value: string("VISA")}

	p.providerdata = data
	p.configured = true
}

func (p *dominosProvider) GetResources(ctx context.Context) (map[string]provider.ResourceType, diag.Diagnostics) {
	return map[string]provider.ResourceType{
		"dominos_order": resourceOrderType{},
	}, nil
}

func (p *dominosProvider) GetDataSources(ctx context.Context) (map[string]provider.DataSourceType, diag.Diagnostics) {
	return map[string]provider.DataSourceType{
		"dominos_address":   dataSourceAddressType{},
		"dominos_store":     dataSourceStoreType{},
		"dominos_menu":      dataSourceMenuType{},
		"dominos_menu_item": dataSourceMenuItemType{},
	}, nil
}

func (p *dominosProvider) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: `
The Dominos provider is used to interact with resources supported by Dominos Pizza.
The provider needs to be configured with a credit card for ordering.

Use the navigation to the right to read about the available resources.
		`,
		Attributes: map[string]tfsdk.Attribute{
			"email_address": {
				Description: "The email address to receive order updates and a receipt to.",
				Required:    true,
				Type:        types.StringType,
			},
			"first_name": {
				Description: "Your first name.",
				Required:    true,
				Type:        types.StringType,
			},
			"last_name": {
				Description: "Your last name.",
				Required:    true,
				Type:        types.StringType,
			},
			"phone_number": {
				Description: "The phone number Dominos will call if any issues arise.",
				Required:    true,
				Type:        types.StringType,
			},
			"credit_card": {
				Description: "Your actual credit card THAT WILL GET CHARGED.",
				Optional:    true,
				Sensitive:   true,
				Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
					"number": {
						Description: "The credit card number.",
						Type:        types.Int64Type,
						Required:    true,
					},
					"cvv": {
						Description: "The credit card CVV.",
						Type:        types.Int64Type,
						Required:    true,
					},
					"date": {
						Description: "The credit card expiration date.",
						Type:        types.StringType,
						Required:    true,
					},
					"postal_code": {
						Description: "The postal code attached to the credit card.",
						Type:        types.StringType,
						Required:    true,
					},
					"card_type": {
						Description: "The credit card type. Default: 'VISA'.",
						Type:        types.StringType,
						Optional:    true,
					},
				}),
			},
		},
	}, nil
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &dominosProvider{
			version: version,
		}
	}
}

// convertProviderType is a helper function for NewResource and NewDataSource
// implementations to associate the concrete provider type. Alternatively,
// this helper can be skipped and the provider type can be directly type
// asserted (e.g. provider: in.(*provider)), however using this can prevent
// potential panics.
func convertProviderType(in provider.Provider) (dominosProvider, diag.Diagnostics) {
	var diags diag.Diagnostics

	p, ok := in.(*dominosProvider)

	if !ok {
		diags.AddError(
			"Unexpected Provider Instance Type",
			fmt.Sprintf("While creating the data source or resource, an unexpected provider type (%T) was received. This is always a bug in the provider code and should be reported to the provider developers.", p),
		)
		return dominosProvider{}, diags
	}

	if p == nil {
		diags.AddError(
			"Unexpected Provider Instance Type",
			"While creating the data source or resource, an unexpected empty provider instance was received. This is always a bug in the provider code and should be reported to the provider developers.",
		)
		return dominosProvider{}, diags
	}

	return *p, diags
}
