package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ provider.DataSourceType = dataSourceAddressType{}
var _ datasource.DataSource = dataSourceAddress{}

type dataSourceAddressType struct{}

func (t dataSourceAddressType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: `
This data source takes in the delivery address and writes it back out in the two different JSON formats that the API expects.

For carryout, this is purely to find the closest store.
		`,
		Attributes: map[string]tfsdk.Attribute{
			"street": {
				Description: "The street to deliver the pizza to. Ex: '123 Main St'.",
				Type:        types.StringType,
				Required:    true,
			},
			"city": {
				Description: "The city to deliver the pizza to. Ex: 'Anytown'.",
				Type:        types.StringType,
				Required:    true,
			},
			"region": {
				Description: "The region to deliver the pizza to, meaning the province or state. Ex: 'BC'.",
				Type:        types.StringType,
				Required:    true,
			},
			"postal_code": {
				Description: "The region to deliver the pizza to (or zip for the USA). Ex: 'A1A1A1'.",
				Type:        types.StringType,
				Required:    true,
			},
			"type": {
				Description: "The type of location to deliver to. Default: 'House'.",
				Type:        types.StringType,
				Optional:    true,
			},
			"url_object": {
				Description: "The computed line1 & line2 for the specified address.",
				Type:        types.StringType,
				Computed:    true,
			},
			"api_object": {
				Description: "The computed json payload for the specified address.",
				Type:        types.StringType,
				Computed:    true,
			},
		},
	}, nil
}

func (t dataSourceAddressType) NewDataSource(ctx context.Context, in provider.Provider) (datasource.DataSource, diag.Diagnostics) {
	provider, diags := convertProviderType(in)

	return dataSourceAddress{
		provider: provider,
	}, diags
}

type dataSourceAddressData struct {
	Street     types.String `tfsdk:"street"`
	City       types.String `tfsdk:"city"`
	Region     types.String `tfsdk:"region"`
	PostalCode types.String `tfsdk:"postal_code"`
	Type       types.String `tfsdk:"type"`
	APIObject  types.String `tfsdk:"api_object"`
	URLObject  types.String `tfsdk:"url_object"`
}

type dataSourceAddress struct {
	provider dominosProvider
}

func (d dataSourceAddress) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data dataSourceAddressData

	diags := req.Config.Get(ctx, &data)
	data.Type = types.String{Value: "House"}

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	urlobj := map[string]string{
		"line1": data.Street.Value,
		"line2": fmt.Sprintf("%s, %s %s", data.City.Value, data.Region.Value, data.PostalCode.Value),
	}
	apiobj := map[string]string{
		"Street":     data.Street.Value,
		"City":       data.City.Value,
		"Region":     data.Region.Value,
		"PostalCode": data.PostalCode.Value,
		"Type":       data.Type.Value,
	}
	url_json, err := json.Marshal(urlobj)
	if err != nil {
		resp.Diagnostics.AddError("Cannot unmarshall urlobj", fmt.Sprintf("%s", err))
		return
	}

	data.URLObject = types.String{Value: string(url_json)}

	api_json, err := json.Marshal(apiobj)
	if err != nil {
		resp.Diagnostics.AddError("Cannot unmarshall apiobj", fmt.Sprintf("%s", err))
		return
	}

	data.APIObject = types.String{Value: string(api_json)}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}
