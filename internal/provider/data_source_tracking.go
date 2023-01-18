package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ provider.DataSourceType = dataSourceTrackingType{}
var _ datasource.DataSource = dataSourceTracking{}

type dataSourceTrackingType struct{}

func (t dataSourceTrackingType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: `
Track a Dominos order.
		`,
		Attributes: map[string]tfsdk.Attribute{
			"store_id": {
				Description: "The ID of the store that the order is for.",
				Type:        types.Int64Type,
				Required:    true,
			},
			"order_id": {
				Description: "The order ID to track.",
				Type:        types.Int64Type,
				Required:    true,
			},
		},
	}, nil
}

func (t dataSourceTrackingType) NewDataSource(ctx context.Context, in provider.Provider) (datasource.DataSource, diag.Diagnostics) {
	provider, diags := convertProviderType(in)

	return dataSourceTracking{
		provider: provider,
	}, diags
}

type dataSourceTrackingData struct {
	StoreID types.Int64 `tfsdk:"store_id"`
	OrderID types.Int64 `tfsdk:"order_id"`
}

type dataSourceTracking struct {
	provider dominosProvider
}

func (d dataSourceTracking) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data dataSourceTrackingData

	diags := req.Config.Get(ctx, &data)

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	var client = &http.Client{Timeout: 10 * time.Second}

	_, err := getTrackingApiObject(fmt.Sprintf("https://trkweb.dominos.ca/orderstorage/GetTrackerData?StoreID=%d&OrderKey=%d", data.StoreID.Value, data.OrderID.Value), client)
	if err != nil {
		resp.Diagnostics.AddError("Cannot get tracking api object", fmt.Sprintf("%s", err))
		return
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func getTrackingApiObject(url string, client *http.Client) (map[string]interface{}, error) {
	r, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	resp := make(map[string]interface{})
	err = json.NewDecoder(r.Body).Decode(&resp)
	return resp, err
}
