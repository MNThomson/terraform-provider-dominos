package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ tfsdk.DataSourceType = dataSourceMenuType{}
var _ tfsdk.DataSource = dataSourceMenu{}

type dataSourceMenuType struct{}

func (t dataSourceMenuType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "Example data source",
		Attributes: map[string]tfsdk.Attribute{
			"store_id": {
				Type:     types.Int64Type,
				Required: true,
			},
			"menu": {
				Computed: true,
				Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
					"name": {
						Type:     types.StringType,
						Computed: true,
					},
					"code": {
						Type:     types.StringType,
						Computed: true,
					},
					"price_cents": {
						Type:     types.Int64Type,
						Computed: true,
					},
				}),
				Description: "Menu Items",
			},
		},
	}, nil
}

func (t dataSourceMenuType) NewDataSource(ctx context.Context, in tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
	provider, diags := convertProviderType(in)

	return dataSourceMenu{
		provider: provider,
	}, diags
}

type dataSourceMenuData struct {
	StoreID types.Int64 `tfsdk:"store_id"`
	Menu    []menuItem  `tfsdk:"delivery_minutes"`
}

type dataSourceMenu struct {
	provider provider
}

func (d dataSourceMenu) Read(ctx context.Context, req tfsdk.ReadDataSourceRequest, resp *tfsdk.ReadDataSourceResponse) {
	var data dataSourceMenuData

	diags := req.Config.Get(ctx, &data)

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	var client = &http.Client{Timeout: 10 * time.Second}

	menuitems, err := getAllMenuItems(fmt.Sprintf("https://order.dominos.com/power/store/%d/menu?lang=en&structured=true", data.StoreID.Value), client)
	if err != nil {
		log.Fatalf("Cannot get all menu items: ", err)
	}

	for i := range menuitems {
		data.Menu = append(data.Menu, menuItem{Name: menuitems[i].Name, Code: menuitems[i].Code, PriceCents: menuitems[i].PriceCents})
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func getMenuApiObject(url string, client *http.Client) (map[string]interface{}, error) {
	r, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	resp := make(map[string]interface{})
	err = json.NewDecoder(r.Body).Decode(&resp)
	return resp, err
}

func getAllMenuItems(url string, client *http.Client) ([]menuItem, error) {
	resp, err := getMenuApiObject(url, client)
	if err != nil {
		return nil, err
	}
	products := resp["Variants"].(map[string]interface{})
	all_products := make([]menuItem, 0, len(products))
	for name, d := range products {
		dict := d.(map[string]interface{})
		price := dict["Price"].(string)
		price = strings.Replace(price, ".", "", 1)
		price_cents, err := strconv.ParseInt(price, 10, 64)
		if err != nil {
			continue
		}
		all_products = append(all_products, menuItem{
			Code:       name,
			Name:       dict["Name"].(string),
			PriceCents: price_cents,
		})
	}
	sort.Slice(all_products, func(i, j int) bool {
		return all_products[i].Code < all_products[j].Code
	})
	// for each entry in Products, make a MenuItem struct and return it.
	return all_products, nil
}
