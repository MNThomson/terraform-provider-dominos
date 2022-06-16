package provider

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ tfsdk.DataSourceType = dataSourceMenuItemType{}
var _ tfsdk.DataSource = dataSourceMenuItem{}

type dataSourceMenuItemType struct{}

func (t dataSourceMenuItemType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "Example data source",
		Attributes: map[string]tfsdk.Attribute{
			"store_id": {
				Type:     types.Int64Type,
				Required: true,
			},
			"query_string": {
				Type: types.ListType{
					ElemType: types.StringType,
				},
				Required: true,
			},
			"matches": {
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
			},
		},
	}, nil
}

func (t dataSourceMenuItemType) NewDataSource(ctx context.Context, in tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
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
	provider provider
}

func (d dataSourceMenuItem) Read(ctx context.Context, req tfsdk.ReadDataSourceRequest, resp *tfsdk.ReadDataSourceResponse) {
	var data dataSourceMenuItemData

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
	/*
			queries := data.QueryString.([]interface{})
		Menu:
			for i := range menuitems {
				for j := range queries {
					if !strings.Contains(strings.ToLower(menuitems[i].Name), strings.ToLower(queries[j].(string))) {
						continue Menu
					}
				}
				data.Matches = append(data.Matches, menuItem{Name: menuitems[i].Name, Code: menuitems[i].Code, PriceCents: menuitems[i].PriceCents})
			}
	*/
	data.Matches = append(data.Matches, menuitems[0])

	log.Printf("len menu: %d", len(data.Matches))

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}
