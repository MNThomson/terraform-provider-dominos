package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ provider.DataSourceType = dataSourceStoreType{}
var _ datasource.DataSource = dataSourceStore{}

type dataSourceStoreType struct{}

func (t dataSourceStoreType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: `
Provided a Dominos address, this data source returns the store_id of the closest Dominos store, and, in case it's useful to you somehow, the delivery_minutes, an integer showing the estimated minutes until your pizza will be delivered.
		`,
		Attributes: map[string]tfsdk.Attribute{
			"address_url_object": {
				Description: "The required line1 & line2 for the specified address.",
				Type:        types.StringType,
				Required:    true,
			},
			"store_id": {
				Description: "The ID of the store closest to the address.",
				Type:        types.Int64Type,
				Computed:    true,
			},
			"delivery_minutes": {
				Description: "The estimated minutes until your pizza will be delivered.",
				Type:        types.Int64Type,
				Computed:    true,
			},
		},
	}, nil
}

func (t dataSourceStoreType) NewDataSource(ctx context.Context, in provider.Provider) (datasource.DataSource, diag.Diagnostics) {
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
	provider dominosProvider
}

func (d dataSourceStore) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
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
		resp.Diagnostics.AddError("Cannot unmarshall address_url_obj", fmt.Sprintf("%s", err))
		return
	}
	line1 := url.QueryEscape(address_url_obj["line1"])
	line2 := url.QueryEscape(address_url_obj["line2"])
	stores, err := getStores(fmt.Sprintf("https://order.dominos.ca/power/store-locator?s=%s&c=%s&s=Delivery", line1, line2), client)
	if err != nil {
		resp.Diagnostics.AddError("Cannot get stores", fmt.Sprintf("%s", err))
		return
	}
	if len(stores) == 0 {
		resp.Diagnostics.AddError("No stores near the address", fmt.Sprintf("%s", err))
		return
	}
	storeID, _ := strconv.ParseInt(stores[0].StoreID, 10, 64)
	data.StoreID = types.Int64{Value: storeID}

	data.DeliveryMinutes = types.Int64{Value: int64(stores[0].ServiceMethodEstimatedWaitMinutes.Delivery.Min)}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

type StoresResponse struct {
	Stores []Store
}

type Store struct {
	StoreID                           string
	ServiceMethodEstimatedWaitMinutes struct {
		Delivery struct {
			Min int64
		}
	}
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
