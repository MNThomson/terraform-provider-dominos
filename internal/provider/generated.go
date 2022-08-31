package provider

import (
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// TODO: Implement https://github.com/hashicorp/terraform-plugin-framework-validators
func pizzaOptionsAttributes() map[string]tfsdk.Attribute {
	tmp := &PizzaOptions{}
	pizzaOptions := make(map[string]tfsdk.Attribute)

	val := reflect.ValueOf(tmp).Elem()
	for i := 0; i < val.NumField(); i++ {
		pizzaOptions[val.Type().Field(i).Name] = tfsdk.Attribute{
			Description: val.Type().Field(i).Name + " DESC",
			Optional:    true,
			Attributes:  itemOptionsAttributes,
		}
	}

	return pizzaOptions
}

type PizzaOptions struct {
	Cheese *Option `json:"C,omitempty"`

	// Pizza Sauces
	PizzaSauce          *Option `json:"X,omitempty"`
	BBQSauce            *Option `json:"Q,omitempty"`
	AlfredoSauce        *Option `json:"Xf,omitempty"`
	HeartyMarinaraSauce *Option `json:"Xm,omitempty"`
	RanchDressing       *Option `json:"Rd,omitempty"`
	GarlicParmesanSauce *Option `json:"Xw,omitempty"`

	// Meats
	Bacon             *Option `json:"K,omitempty"`
	BeefCrumble       *Option `json:"B,omitempty"`
	BrooklynPepperoni *Option `json:"Xp,omitempty"`
	Chicken           *Option `json:"D,omitempty"`
	Ham               *Option `json:"H,omitempty"`
	Pepperoni         *Option `json:"P,omitempty"`
	PhillySteak       *Option `json:"St,omitempty"`
	Salami            *Option `json:"L,omitempty"`
	Sausage           *Option `json:"S,omitempty"`

	// Non-meats
	BabySpinach       *Option `json:"Sp,omitempty"`
	BlackOlives       *Option `json:"R,omitempty"`
	Cheddar           *Option `json:"E,omitempty"`
	Feta              *Option `json:"Fe,omitempty"`
	GreenOlives       *Option `json:"V,omitempty"`
	GreenPepper       *Option `json:"G,omitempty"`
	HotBananaPeppers  *Option `json:"Z,omitempty"`
	JalapenoPeppers   *Option `json:"J,omitempty"`
	Mushroom          *Option `json:"M,omitempty"`
	Onion             *Option `json:"O,omitempty"`
	ParmesanAsiago    *Option `json:"Pa,omitempty"`
	Pineapple         *Option `json:"N,omitempty"`
	Provolone         *Option `json:"Cp,omitempty"`
	RoastedRedPeppers *Option `json:"Rp,omitempty"`
	Tomatoes          *Option `json:"T,omitempty"`
}
