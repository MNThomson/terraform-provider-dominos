package provider

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ provider.DataSourceType = dataSourcePizzaType{}
var _ datasource.DataSource = dataSourcePizza{}

type dataSourcePizzaType struct{}

var itemOptionsAttributes = tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
	"portion": {
		Description: "Choose:",
		Optional:    true,
		Type:        types.StringType,
	},
	"weight": {
		Description: "Choose:",
		Optional:    true,
		Type:        types.StringType,
	},
})

func (t dataSourcePizzaType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: `
Provided a Dominos address, this data source returns the store_id of the closest Dominos store, and, in case it's useful to you somehow, the delivery_minutes, an integer showing the estimated minutes until your pizza will be delivered.
		`,
		Attributes: map[string]tfsdk.Attribute{
			"size": {
				Description: "",
				Type:        types.StringType,
				Required:    true,
			},
			"crust": {
				Description: "",
				Type:        types.StringType,
				Required:    true,
			},
			"options": {
				Description: "",
				Optional:    true,
				Attributes:  tfsdk.SingleNestedAttributes(pizzaOptionsAttributes()),
			},
			"pizza_json": {
				Description: "The json for the pizza Product.",
				Type:        types.StringType,
				Computed:    true,
			},
			"quantity": {
				Description: "",
				Type:        types.NumberType,
				Optional:    true,
			},
		},
	}, nil
}

func (t dataSourcePizzaType) NewDataSource(ctx context.Context, in provider.Provider) (datasource.DataSource, diag.Diagnostics) {
	provider, diags := convertProviderType(in)

	return dataSourcePizza{
		provider: provider,
	}, diags
}

type dataSourcePizzaData struct {
	Size      types.String `tfsdk:"size"`
	Crust     types.String `tfsdk:"crust"`
	Options   types.Object `tfsdk:"options"`
	PizzaJson types.String `tfsdk:"pizza_json"`
	Quantity  types.Number `tfsdk:"quantity"`
}

type dataSourcePizza struct {
	provider dominosProvider
}

func (d dataSourcePizza) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data dataSourcePizzaData
	var PizzaJson Product

	diags := req.Config.Get(ctx, &data)

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// for key, val := range data.Options.Attrs {
	// 	fmt.Printf("Key: %s, Value: %s\n", key, val)
	// }

	// Pizza Code
	// TODO: Validate crust is offered
	PizzaJson.Code = data.Size.Value + data.Crust.Value
	if data.Quantity.Null {
		PizzaJson.Qty = 1
	} else {
		qty, _ := data.Quantity.Value.Int64()
		PizzaJson.Qty = int(qty)
	}

	out, _ := json.Marshal(PizzaJson)
	output := string(out)

	tflog.Info(ctx, string(out))

	data.PizzaJson = types.String{Value: output}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}
