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
				Type:     types.StringType,
				Required: true,
			},
			"menu": {
				Type: types.ListType{
					ElemType: types.StringType,
				},
				Computed: true,
				Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
					"name": {
						Type:     types.Int64Type,
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
	StoreID         types.Int64 `tfsdk:"store_id"`
	DeliveryMinutes types.Int64 `tfsdk:"delivery_minutes"`
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

	menuitems, err := getAllMenuItems(fmt.Sprintf("https://order.dominos.com/power/store/%s/menu?lang=en&structured=true", data.StoreID), client)
	if err != nil {
		log.Fatalf("Cannot get all menu items")
	}
	menu := make([]map[string]interface{}, 0, len(menuitems))
	for i := range menuitems {
		menu = append(menu, map[string]interface{}{"name": menuitems[i].Name, "code": menuitems[i].Code, "price_cents": menuitems[i].PriceCents})
	}

	log.Printf("len menu: %d", len(menu))
	data.StoreID.Value, _ = strconv.ParseInt(stores[0].StoreID, 10, 64)
	data.StoreID.Null = false //TODO: Set properly

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

/*
func resourceMenuRead(d *schema.ResourceData, m interface{}) error {
	var client = &http.Client{Timeout: 10 * time.Second}
	store_id := d.Get("store_id").(string)
	menuitems, err := getAllMenuItems(fmt.Sprintf("https://order.dominos.com/power/store/%s/menu?lang=en&structured=true", store_id), client)
	if err != nil {
		return err
	}
	menu := make([]map[string]interface{}, 0, len(menuitems))
	for i := range menuitems {
		menu = append(menu, map[string]interface{}{"name": menuitems[i].Name, "code": menuitems[i].Code, "price_cents": menuitems[i].PriceCents})
	}
	if err := d.Set("menu", menu); err != nil {
		return err
	}
	log.Printf("len menu: %d", len(menu))
	d.SetId(store_id)
	return nil
}
*/
type MenuItem struct {
	Code       string
	Name       string
	PriceCents int64
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

func getAllMenuItems(url string, client *http.Client) ([]MenuItem, error) {
	resp, err := getMenuApiObject(url, client)
	if err != nil {
		return nil, err
	}
	products := resp["Variants"].(map[string]interface{})
	all_products := make([]MenuItem, 0, len(products))
	log.Printf("len products: %d", len(products))
	for name, d := range products {
		dict := d.(map[string]interface{})
		price := dict["Price"].(string)
		price = strings.Replace(price, ".", "", 1)
		price_cents, err := strconv.ParseInt(price, 10, 64)
		if err != nil {
			continue
		}
		all_products = append(all_products, MenuItem{
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
