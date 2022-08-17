package provider

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ provider.DataSourceType = dataSourceMenuItemType{}
var _ datasource.DataSource = dataSourceMenuItem{}

type dataSourceMenuItemType struct{}

func (t dataSourceMenuItemType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: `
This data source takes in the store_id and a list of strings (as query_string), and outputs the menu items in matches.
Each item in matches has three attributes: name, code, and price_cents.
The name is human-readable, but not useful for ordering.
The price_cents is also only informational.

Each string in query_string must literally match the name of the menu item for the menu item to appear in matches.
		`,
		Attributes: map[string]tfsdk.Attribute{
			"store_id": {
				Description: "The ID of the store to get the menu for.",
				Type:        types.Int64Type,
				Required:    true,
			},
			"query_string": {
				Description: "Each string in query_string must literally match the name of the menu item for the menu item to appear in matches.",
				Type: types.ListType{
					ElemType: types.StringType,
				},
				Required: true,
			},
			"matches": {
				Description: "An array of all possible menu item that matches the given query string.",
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

func (t dataSourceMenuItemType) NewDataSource(ctx context.Context, in provider.Provider) (datasource.DataSource, diag.Diagnostics) {
	provider, diags := convertProviderType(in)

	return dataSourceMenuItem{
		provider: provider,
	}, diags
}

type dataSourceMenuItemData struct {
	StoreID     types.Int64    `tfsdk:"store_id"`
	QueryString []types.String `tfsdk:"query_string"`
	Matches     []menuItem     `tfsdk:"matches"`
}

type menuItem struct {
	Name       string `tfsdk:"name"`
	Code       string `tfsdk:"code"`
	PriceCents int64  `tfsdk:"price_cents"`
}

type dataSourceMenuItem struct {
	provider dominosProvider
}

func (d dataSourceMenuItem) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data dataSourceMenuItemData

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

	queries := data.QueryString

NextItem:
	for i := range menuitems {
		for j := range queries {
			// Remove quotes from query string
			query := queries[j].String()[1 : len(queries[j].String())-1]

			if !strings.Contains(strings.ToLower(menuitems[i].Name), strings.ToLower(query)) {
				continue NextItem
			}
		}
		data.Matches = append(data.Matches, menuItem{Name: menuitems[i].Name, Code: menuitems[i].Code, PriceCents: menuitems[i].PriceCents})
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}
