package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
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
	var pizzaJson Product

	diags := req.Config.Get(ctx, &data)

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Pizza Code
	// TODO: Validate crust is offered
	pizzaJson.Code = data.Size.Value + data.Crust.Value
	if data.Quantity.Null {
		pizzaJson.Qty = 1
	} else {
		qty, _ := data.Quantity.Value.Int64()
		pizzaJson.Qty = int(qty)
	}

	if !data.Options.IsNull() {
		var tmp1 TFPizzaOption
		mapTest := make(map[string]Option)

		for optionName, optionVal := range data.Options.Attrs {
			pizzaOptionMap := make(map[string]string)
			var key string
			var weight string

			optionValStr := strings.ReplaceAll(optionVal.String(), "<null>", "null")

			if optionVal.IsNull() {
				continue
			}

			err := json.Unmarshal([]byte(optionValStr), &tmp1)
			if err != nil {
				resp.Diagnostics.AddError("Cannot unmarshall Stuff", fmt.Sprintf("%s", err))
				return
			}

			if tmp1.Portion != nil {
				switch *tmp1.Portion {
				case "left":
					key = "left"
				case "all":
					key = "all"
				case "right":
					key = "right"
				default:
					resp.Diagnostics.AddError("Portion not valid:", fmt.Sprintf("%s", *tmp1.Weight))
					return
				}
			} else {
				key = "all"
			}

			if tmp1.Weight != nil {
				switch *tmp1.Weight {
				case "light":
					weight = "0.5"
				case "normal":
					weight = "1"
				case "extra":
					weight = "1.5"
				default:
					resp.Diagnostics.AddError("Weight not valid:", fmt.Sprintf("%s", *tmp1.Weight))
					return
				}
			} else {
				weight = "1"
			}

			pizzaOptionMap[key] = weight
			var pizzaOption Option
			mapstructure.Decode(pizzaOptionMap, &pizzaOption)

			mapTest[optionName] = pizzaOption
		}

		mapstructure.Decode(mapTest, &pizzaJson.Options)
	}

	out, _ := json.Marshal(pizzaJson)
	output := string(out)

	tflog.Info(ctx, string(out))

	data.PizzaJson = types.String{Value: output}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}
