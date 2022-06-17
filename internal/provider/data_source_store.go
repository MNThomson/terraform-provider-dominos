package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ tfsdk.DataSourceType = dataSourceStoreType{}
var _ tfsdk.DataSource = dataSourceStore{}

type dataSourceStoreType struct{}

func (t dataSourceStoreType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "Example data source",
		Attributes: map[string]tfsdk.Attribute{
			"address_url_object": {
				Type:     types.StringType,
				Required: true,
			},
			"store_id": {
				Type:     types.Int64Type,
				Computed: true,
			},
			"delivery_minutes": {
				Type:     types.Int64Type,
				Computed: true,
			},
		},
	}, nil
}

func (t dataSourceStoreType) NewDataSource(ctx context.Context, in tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
	provider, diags := convertProviderType(in)

	return dataSourceStore{
		provider: provider,
	}, diags
}

type dataSourceStoreData struct {
	AddressURLObj   types.String `tfsdk:"address_url_object"`
	StoreID         types.Int64  `tfsdk:"store_id"`
	DeliveryMinutes types.Int64  `tfsdk:"delivery_minutes"`
}

type dataSourceStore struct {
	provider provider
}

func (d dataSourceStore) Read(ctx context.Context, req tfsdk.ReadDataSourceRequest, resp *tfsdk.ReadDataSourceResponse) {
	var data dataSourceStoreData

	diags := req.Config.Get(ctx, &data)

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	var client = &http.Client{Timeout: 10 * time.Second}
	address_url_obj := make(map[string]string)
	err := json.Unmarshal([]byte(data.AddressURLObj.Value), &address_url_obj)
	if err != nil {
		log.Fatalf("Cannot unmarshall address_url_obj")
	}
	line1 := url.QueryEscape(address_url_obj["line1"])
	line2 := url.QueryEscape(address_url_obj["line2"])
	stores, err := getStores(fmt.Sprintf("https://order.dominos.com/power/store-locator?s=%s&c=%s&s=Delivery", line1, line2), client)
	if err != nil {
		log.Fatalf("Cannot get stores")
	}
	if len(stores) == 0 {
		log.Fatalf("No stores near the address %#v", address_url_obj)
	}

	data.StoreID.Value, _ = strconv.ParseInt(stores[0].StoreID, 10, 64)
	data.StoreID.Null = false //TODO: Set properly

	data.DeliveryMinutes.Value = int64(stores[0].ServiceMethodEstimatedWaitMinutes.Delivery.Min)
	data.DeliveryMinutes.Null = false //TODO: Set properly

	// d.SetId("store")

	fmt.Println("\n", data.StoreID.Value, "\nTEST1")

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

type StoresResponse struct {
	Stores []Store
}

type Store struct {
	StoreID                           string
	ServiceMethodEstimatedWaitMinutes WaitMinutes
}

type WaitMinutes struct {
	Delivery DeliveryMinutes
}

type DeliveryMinutes struct {
	Min int
}

func getStores(url string, client *http.Client) ([]Store, error) {
	r, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	resp := StoresResponse{}

	err = json.NewDecoder(r.Body).Decode(&resp)
	if err != nil {
		return nil, err
	}
	return resp.Stores, nil
}
