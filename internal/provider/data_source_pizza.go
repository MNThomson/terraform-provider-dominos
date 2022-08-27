package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ provider.DataSourceType = dataSourcePizzaType{}
var _ datasource.DataSource = dataSourcePizza{}

type dataSourcePizzaType struct{}

var itemOptionsAttributes = tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
	"portion": {
		Description: "",
		Optional:    true,
		Type:        types.StringType,
	},
	"weight": {
		Description: "",
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
				Attributes:  optionsAttributes,
			},
			"pizza_json": {
				Description: "The json for the pizza Product.",
				Type:        types.StringType,
				Computed:    true,
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
}

type dataSourcePizza struct {
	provider dominosProvider
}

func (d dataSourcePizza) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data dataSourcePizzaData

	diags := req.Config.Get(ctx, &data)

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// for key, val := range data.Options.Attrs {
	// 	fmt.Printf("Key: %s, Value: %s\n", key, val)
	// }

	data.PizzaJson = types.String{Value: data.Options.String()}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}
