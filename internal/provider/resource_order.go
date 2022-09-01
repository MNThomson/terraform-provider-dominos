package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/mnthomson/terraform-provider-dominos/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ provider.ResourceType = resourceOrderType{}
var _ resource.Resource = resourceOrder{}
var _ resource.ResourceWithImportState = resourceOrder{}

type resourceOrderType struct{}

func (t resourceOrderType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		// This description is used by the documentation generator and the language server.
		Description: `
This is it! This will order you your pizzas!

As far as I know there is no way to cancel a dominos order programmatically, so if you made a mistake, you'll have to call the store.
You should receive an email confirmation almost instantly, and that email will have the store's phone number in it.
		`,
		Attributes: map[string]tfsdk.Attribute{
			"address_api_object": {
				Description: "The computed json payload for the specified address.",
				Required:    true,
				Type:        types.StringType,
			},
			"item_codes": {
				Description: "An array of menu items to order.",
				Required:    true,
				Type: types.ListType{
					ElemType: types.StringType,
				},
			},
			"store_id": {
				Description: "The ID of the store that the order is for.",
				Required:    true,
				Type:        types.Int64Type,
			},
			"total_price": {
				Description: "The computed total price of the order.",
				Computed:    true,
				Type:        types.NumberType,
			},
		},
	}, nil
}

func (t resourceOrderType) NewResource(ctx context.Context, in provider.Provider) (resource.Resource, diag.Diagnostics) {
	provider, diags := convertProviderType(in)

	return resourceOrder{
		provider: provider,
	}, diags
}

type resourceOrderData struct {
	AddressAPIObj types.String `tfsdk:"address_api_object"`
	ItemCodes     types.List   `tfsdk:"item_codes"`
	StoreID       types.Int64  `tfsdk:"store_id"`
	TotalPrice    types.Number `tfsdk:"total_price"`
}

type resourceOrder struct {
	provider dominosProvider
}

type Address struct {
	Street               string `json:"Street"`
	City                 string `json:"City"`
	Region               string `json:"Region"`
	PostalCode           string `json:"PostalCode"`
	Type                 string `json:"Type" default:"House"`
	DeliveryInstructions string `json:"DeliveryInstructions"`
}

type Payment struct {
	Type         string  `json:"Type" default:"DoorCredit"`
	Amount       float64 `json:"Amount"`
	Number       string  `json:"Number" default:""`
	CardType     string  `json:"CardType" default:""`
	Expiration   string  `json:"Expiration" default:""`
	SecurityCode string  `json:"SecurityCode" default:""`
	PostalCode   string  `json:"PostalCode" default:""`
	ProviderID   string  `json:"ProviderID" default:""`
	// PaymentMethodID string `json:"PaymentMethodID"`
	// OTP string `json:"OTP"`
	// GpmPaymentType string `json:"gpmPaymentType"`
}

type Product struct {
	Code                 string       `json:"Code"`
	Qty                  int          `json:"Qty" default:"1"`
	ID                   int          `json:"ID"`
	IsNew                bool         `json:"isNew" default:"true"`
	ShowBestPriceMessage bool         `json:"ShowBestPriceMessage" default:"false"`
	Options              PizzaOptions `json:"Options"`
}

/*
 * Light: "0.5"
 * Normal: "1"
 * Extra: "1.5"
 */
type Option struct {
	Left  string `json:"1/2,omitempty" mapstructure:"left"`
	All   string `json:"1/1,omitempty" mapstructure:"all"`
	Right string `json:"2/2,omitempty" mapstructure:"right"`
}

type TFPizzaOption struct {
	Portion *string `json:"portion,omitempty"`
	Weight  *string `json:"weight,omitempty"`
}

type DominosOrderData struct {
	Order struct {
		Address               Address    `json:"Address"`
		Coupons               []struct{} `json:"Coupons"`
		CustomerID            string     `json:"CustomerID" default:""`
		Email                 string     `json:"Email"`
		Extension             string     `json:"Extension" default:""`
		FirstName             string     `json:"FirstName"`
		LastName              string     `json:"LastName"`
		LanguageCode          string     `json:"LanguageCode" default:"en"`
		OrderChannel          string     `json:"OrderChannel" default:"OLO"`
		OrderID               string     `json:"OrderID" default:""`
		OrderMethod           string     `json:"OrderMethod" default:"Web"`
		OrderTaker            struct{}   `json:"OrderTaker"`
		Payments              []Payment  `json:"Payments"`
		Phone                 string     `json:"Phone"`
		PhonePrefix           string     `json:"PhonePrefix" default:""`
		Products              []Product  `json:"Products"`
		ServiceMethod         string     `json:"ServiceMethod" default:"Delivery"`
		SourceOrganizationURI string     `json:"SourceOrganizationURI" default:"order.dominos.com"`
		StoreID               string     `json:"StoreID"`
		Tags                  struct{}   `json:"Tags"`
		Version               string     `json:"Version" default:"1.0"`
		NoCombine             bool       `json:"NoCombine" default:"true"`
		Partners              struct{}   `json:"Partners"`
		HotspotsLite          bool       `json:"HotspotsLite" default:"false"`
		OrderInfoCollection   []struct{} `json:"OrderInfoCollection"`
		NewUser               bool       `json:"NewUser" default:"true"`
	} `json:"Order"`
}

func price_order(ctx context.Context, order_data DominosOrderData, client *http.Client) (string, error) {
	output_bytes, err := json.Marshal(order_data)
	if err != nil {
		return "", fmt.Errorf("Error Marshalling Order Data: %s", err)
	}

	val_req, err := http.NewRequest("POST", "https://order.dominos.ca/power/price-order", strings.NewReader(string(output_bytes)))
	if err != nil {
		return "", fmt.Errorf("HTTP Error: %s", err)
	}

	val_req.Header.Set("Referer", "https://order.dominos.ca/en/pages/order/")
	val_req.Header.Set("Content-Type", "application/json")

	dumpreq, err := httputil.DumpRequest(val_req, true)
	if err != nil {
		return "", fmt.Errorf("HTTP Error: %s", err)
	}

	tflog.Info(ctx, "http request:\n"+string(dumpreq)+"\n")

	val_rsp, err := client.Do(val_req)
	if err != nil {
		return "", fmt.Errorf("HTTP Error: %s", err)
	}

	dumprsp, err := httputil.DumpResponse(val_rsp, true)
	if err != nil {
		return "", fmt.Errorf("HTTP Error: %s", err)
	}

	tflog.Info(ctx, "http response:\n"+string(dumprsp)+"\n")
	validate_response_obj := make(map[string]interface{})
	err = json.NewDecoder(val_rsp.Body).Decode(&validate_response_obj)
	if err != nil {
		return "", fmt.Errorf("JSON Decoder error: %s", err)
	}

	if validate_response_obj["Status"].(float64) == -1 {
		return "", fmt.Errorf("The Dominos API didn't like this request: %s", validate_response_obj["StatusItems"])
	}
	return "Hello", nil
}

func (r resourceOrder) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data resourceOrderData
	var providerdata providerData
	var client = &http.Client{Timeout: 10 * time.Second}

	diags := req.Config.Get(ctx, &data)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	providerdata = r.provider.providerdata
	order_data := &DominosOrderData{}

	// Order data defaults
	err := utils.Set(&(order_data.Order), "default")
	if err != nil {
		resp.Diagnostics.AddError("Error Setting Order Defaults", fmt.Sprintf("%s", err))
		return
	}

	// Address defaults
	err = utils.Set(&(order_data.Order.Address), "default")
	if err != nil {
		resp.Diagnostics.AddError("Error Setting Address Defaults", fmt.Sprintf("%s", err))
		return
	}

	// Provided address data
	err = json.Unmarshal([]byte(data.AddressAPIObj.Value), &(order_data.Order.Address))
	if err != nil {
		resp.Diagnostics.AddError("Error unmarshalling AddressAPIObj", fmt.Sprintf("%s", err))
		return
	}

	// Provided personal data
	order_data.Order.FirstName = providerdata.FirstName.Value
	order_data.Order.LastName = providerdata.LastName.Value
	order_data.Order.Email = providerdata.EmailAddr.Value
	order_data.Order.Phone = providerdata.PhoneNumber.Value

	// Misc
	order_data.Order.StoreID = strconv.FormatInt(data.StoreID.Value, 10)

	/*
	 * Add items to Products
	 */

	/*
	 * Price Order
	 */
	price_res, err := price_order(ctx, *order_data, client)
	if err != nil {
		resp.Diagnostics.AddError("Error Pricing Order", fmt.Sprintf("%s", err))
		return
	}
	fmt.Println(price_res)

	/*
	 * Order order
	 */
	// Add Payment details

	// Payment defaults
	var payment Payment
	order_data.Order.Payments = append(order_data.Order.Payments, payment)
	err = utils.Set(&(order_data.Order.Payments[0]), "default")
	if err != nil {
		resp.Diagnostics.AddError("Error Setting Payment Defaults", fmt.Sprintf("%s", err))
		return
	}

	// Set price

	//Place Order
	// place_res, err := place_order(*order_data, client)
	// if err != nil {
	// 	return
	// }
	// fmt.Println(place_res)

	//Printing output
	// output_bytes, err := json.Marshal(order_data)
	// if err != nil {
	// 	resp.Diagnostics.AddError("Error Placing Order", fmt.Sprintf("%s", err))
	// 	return
	// }

	// output := string(output_bytes)
	// fmt.Println(output)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r resourceOrder) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data resourceOrderData

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r resourceOrder) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data resourceOrderData

	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r resourceOrder) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data resourceOrderData

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
}

func (r resourceOrder) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
