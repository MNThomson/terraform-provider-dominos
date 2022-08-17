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

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ provider.DataSourceType = dataSourceMenuType{}
var _ datasource.DataSource = dataSourceMenu{}

type dataSourceMenuType struct{}

func (t dataSourceMenuType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: `
If you would prefer to do your own filtering, you can get access to every item on the dominos menu in your area using this data source.
This data source takes in store_id and provides menu, a list of all (186, at my dominos) name/code/price_cents blocks.

For the love of all that's holy, do not accidentally feed this data source directly into the dominos_order.
This will be expensive and probably pretty annoying to the Dominos store, which will be serving you 1 of each 2-liter bottle of soda, 1 of each 20oz bottle, at least 4 different kinds of salad, probably like 6 different kinds of chicken wings, and I think 12 of each kind of pizza?
(Small, medium, large) x (Hand Tossed, Pan, Stuffed Crust, Gluten Free)?
Oh plus breads. There's breads on the menu, I found that out while trawling through API responses.
I wonder who eats those. Are they good? Let me know!
		`,
		Attributes: map[string]tfsdk.Attribute{
			"store_id": {
				Description: "The ID of the store to get the menu for.",
				Type:        types.Int64Type,
				Required:    true,
			},
			"menu": {
				Description: "An array of all menu item for the given store.",
				Computed:    true,
				Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
					"name": {
						Description: "The name of the item.",
						Type:        types.StringType,
						Computed:    true,
					},
					"code": {
						Description: "The dominos code for the item.",
						Type:        types.StringType,
						Computed:    true,
					},
					"price_cents": {
						Description: "The price in cents of the item.",
						Type:        types.Int64Type,
						Computed:    true,
					},
				}),
			},
		},
	}, nil
}

func (t dataSourceMenuType) NewDataSource(ctx context.Context, in provider.Provider) (datasource.DataSource, diag.Diagnostics) {
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
	provider dominosProvider
}

func (d dataSourceMenu) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data dataSourceMenuData

	diags := req.Config.Get(ctx, &data)

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	var client = &http.Client{Timeout: 10 * time.Second}

	menuitems, err := getAllMenuItems(fmt.Sprintf("https://order.dominos.com/power/store/%d/menu?lang=en&structured=true", data.StoreID.Value), client)
	if err != nil {
		log.Fatalf("Cannot get all menu items: %v", err)
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
