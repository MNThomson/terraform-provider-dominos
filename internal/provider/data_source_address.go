package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ tfsdk.DataSourceType = dataSourceAddressType{}
var _ tfsdk.DataSource = dataSourceAddress{}

type dataSourceAddressType struct{}

func (t dataSourceAddressType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Example data source",

		Attributes: map[string]tfsdk.Attribute{
			"street": {
				Type:     types.StringType,
				Required: true,
			},
			"city": {
				Type:     types.StringType,
				Required: true,
			},
			"state": {
				Type:     types.StringType,
				Required: true,
			},
			"zip": {
				Type:     types.StringType,
				Required: true,
			},
			"url_object": {
				Type:     types.StringType,
				Computed: true,
			},
			"api_object": {
				Type:     types.StringType,
				Computed: true,
			},
		},
	}, nil
}

func (t dataSourceAddressType) NewDataSource(ctx context.Context, in tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
	provider, diags := convertProviderType(in)

	return dataSourceAddress{
		provider: provider,
	}, diags
}

type dataSourceAddressData struct {
	Street     types.String `tfsdk:"street"`
	City       types.String `tfsdk:"city"`
	Region     types.String `tfsdk:"state"`
	PostalCode types.String `tfsdk:"zip"`
	Type       types.String `tfsdk:"type"`
	APIObject  types.String `tfsdk:"api_object"`
	URLObject  types.String `tfsdk:"url_object"`
}

type dataSourceAddress struct {
	provider provider
}

func (d dataSourceAddress) Read(ctx context.Context, req tfsdk.ReadDataSourceRequest, resp *tfsdk.ReadDataSourceResponse) {
	var data dataSourceAddressData

	diags := req.Config.Get(ctx, &data)
	data.Type = types.String{Value: "House"}

	resp.Diagnostics.Append(diags...)

	log.Printf("got here")

	if resp.Diagnostics.HasError() {
		return
	}

	log.Printf("got here")

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	// example, err := d.provider.client.ReadExample(...)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read example, got error: %s", err))
	//     return
	// }

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
		log.Fatalf("Cannot unmarshall urlobj")

	}

	url_json_string := string(url_json)
	log.Printf("[DEBUG] url json: %#v to %s", urlobj, url_json_string)

	api_json, err := json.Marshal(apiobj)
	if err != nil {
		log.Fatalf("Cannot unmarshall apiobj")
	}
	api_json_string := string(api_json)
	data.APIObject.Value = api_json_string

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

/*

func resourceAddressRead(d *schema.ResourceData, m interface{}) error {
	d.SetId("address")
	urlobj := map[string]string{
		"line1": d.Get("street").(string),
		"line2": fmt.Sprintf("%s, %s %s", d.Get("city").(string), d.Get("state").(string), d.Get("zip").(string)),
	}
	apiobj := map[string]string{
		"Street":     d.Get("street").(string),
		"City":       d.Get("city").(string),
		"Region":     d.Get("state").(string),
		"PostalCode": d.Get("zip").(string),
		"Type":       "House",
	}
	url_json, err := json.Marshal(urlobj)
	if err != nil {
		return err
	}
	url_json_string := string(url_json)
	log.Printf("[DEBUG] url json: %#v to %s", urlobj, url_json_string)
	if err := d.Set("url_object", url_json_string); err != nil {
		return err
	}
	api_json, err := json.Marshal(apiobj)
	if err != nil {
		return err
	}
	api_json_string := string(api_json)
	if err := d.Set("api_object", api_json_string); err != nil {
		return err
	}
	return nil
}

*/
