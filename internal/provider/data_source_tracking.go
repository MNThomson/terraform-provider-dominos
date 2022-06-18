package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ tfsdk.DataSourceType = dataSourceTrackingType{}
var _ tfsdk.DataSource = dataSourceTracking{}

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

func (t dataSourceTrackingType) NewDataSource(ctx context.Context, in tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
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
	provider provider
}

func (d dataSourceTracking) Read(ctx context.Context, req tfsdk.ReadDataSourceRequest, resp *tfsdk.ReadDataSourceResponse) {
	var data dataSourceTrackingData

	diags := req.Config.Get(ctx, &data)

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	var client = &http.Client{Timeout: 10 * time.Second}

	_, err := getTrackingApiObject(fmt.Sprintf("https://trkweb.dominos.com/orderstorage/GetTrackerData?StoreID=%d&OrderKey=%d", data.StoreID.Value, data.OrderID.Value), client)
	if err != nil {
		log.Fatalf("Cannot get tracking api object: %v", err)
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
