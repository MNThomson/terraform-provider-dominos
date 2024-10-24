# Terraform Provider for Dominos Pizza

[![GitHub Release Workflow Status](https://img.shields.io/github/actions/workflow/status/MNThomson/terraform-provider-dominos/release.yml?label=Build&labelColor=black&logo=GitHub%20Actions&style=flat-square)](https://github.com/MNThomson/terraform-provider-dominos/actions/workflows/release.yml)
[![Terraform Registry Version](https://img.shields.io/github/v/release/MNThomson/terraform-provider-dominos?labelColor=black&label=TF%20Registry&logo=terraform&logoColor=7b42bc&color=7b42bc&style=flat-square)](https://registry.terraform.io/providers/MNThomson/dominos/latest)
[![Terraform Registry Downloads](https://img.shields.io/badge/dynamic/json?color=7b42bc&label=Downloads&labelColor=black&logo=terraform&logoColor=7b42bc&query=data.attributes.total&url=https%3A%2F%2Fregistry.terraform.io%2Fv2%2Fproviders%2F3133%2Fdownloads%2Fsummary&style=flat-square)](https://registry.terraform.io/providers/MNThomson/dominos/latest)

---

The Dominos provider exists to ensure that while your cloud infrastructure is spinning up, you can have a hot pizza delivered to you. This paradigm-shifting expansion of Terraform's "resource" model into the physical world was inspired in part by the realization that Google has a REST API for Interconnects, e.g. for people with hard-hats digging up the ground, laying fiber. If you can use Terraform to summon folks with shovels to drop a fiber line, why shouldn't you be able to summon a driver with a pizza?

## Example Usage

```terraform
terraform {
  required_providers {
    dominos = {
      source = "MNThomson/dominos"
    }
  }
}

provider "dominos" {
  first_name    = "My"
  last_name     = "Name"
  email_address = "my@name.com"
  phone_number  = "15555555555"

  credit_card = {
    number      = 123456789101112
    cvv         = 1314
    date        = "15/16"
    postal_code = "18192"
  }
}

data "dominos_address" "addr" {
  street      = "123 Main St"
  city        = "Anytown"
  region      = "WA"
  postal_code = "02122"
}

data "dominos_store" "store" {
  address_url_object = data.dominos_address.addr.url_object
}

data "dominos_menu_item" "item" {
  store_id     = data.dominos_store.store.store_id
  query_string = ["philly", "medium"]
}

output "OrderOutput" {
  value = data.dominos_menu_item.item.matches[*]
}

resource "dominos_order" "order" {
  api_object = data.dominos_address.addr.api_object
  item_codes = data.dominos_menu_item.item.matches[*].code
  store_id   = data.dominos_store.store.store_id
}
```

As usual, `terraform init` and `terraform plan`! Run `terraform apply` when ready - but use caution, since this is going to charge you money.

Now I don't know what you're going to get since I don't know what a medium philly is in your area, but in my area it gets you a 12" hand-tossed philly cheesesteak pizza, and it's pretty good. It's all right. Regular Dominos.

## Provider Overview

The Dominos Pizza provider is made up primarily of data sources. The only thing you can truly `Create` with this provider is, of course, an order from Dominos.

If you are a true Dominos aficionado, you may already know the four-digit store ID of the store closest to you, the correct json-format for your address, the six-to-ten-digit code for the item you want to order. If you are one of those people, you can feel free to construct a `dominos_order` resource from scratch.

For the rest of us, I recommend one of each of the data sources. They feed into each other in an obvious way.

## Credit

Massive credit to [nat-henderson](https://github.com/nat-henderson/terraform-provider-dominos): they built the kitchen, assembled the wood fired oven, and perfected the recipe. I am merely the waiter serving this pizza to the masses.

## Warnings and Caveats

1) The author(s) of this software are not in any sense associated with Dominos Pizza. It was an idea a bunch of us had while working on the Google provider, but this software isn't associated with Google, either. For further details you can read LICENSE.md.

2) If your cloud infrastructure is slow to spin up, your pizza might arrive before your changes finish applying. This will be embarrassing, and potentially distracting.

3) This is not a joke provider. Or, it kind of is a joke, but even though it's a joke it will still order you a pizza. You are going to get a pizza. You should be careful with this provider, if you don't want a pizza.

4) Even if you do want a pizza, you should probably be careful with this provider. In testing, I once nearly ordered every item on the Domino's menu, which would probably have been expensive and embarrassing.

5) You do have to put your actual credit card information into this provider, because you will, again, be purchasing and receiving a pizza.

6) Although all your credit card information is marked `Sensitive` in schema, that's the only protection they've got. If your state storage isn't secure, maybe don't use this provider. Or use a virtual card number, or COD, or something. Be smart. Again, real credit card, real money, real pizza.

7) I cannot emphasize enough how much you are actually going to be ordering a pizza. Please do not be surprised when you receive a pizza and a corresponding charge to your credit card.

8) As far as I know, there is no programmatic way to `destroy` an existing pizza. `terraform destroy` is implemented on the client side, by consuming the pizza.

9) The Dominos API supports an astonishing amount of customization of your items. I think this is where "none pizza with left beef" comes from. You can't do any of that with this provider. Order off the menu!

10) Dominos probably exists outside the US, but I have no idea what will happen if you try to order a pizza outside the US. Some quick testing suggests it just times out.

11) This provider auto-accepts Dominos' canonicalization of your address. If you live someplace the post office doesn't know about, you might have trouble.

# Pendantic Schema

<details>
  <summary>Expand</summary>

    <!-- schema generated by tfplugindocs -->
## Schema

### Required

- `email_address` (String) The email address to receive order updates and a receipt to.
- `first_name` (String) Your first name.
- `last_name` (String) Your last name.
- `phone_number` (String) The phone number Dominos will call if any issues arise.

### Optional

- `credit_card` (Attributes, Sensitive) Your actual credit card THAT WILL GET CHARGED. (see [below for nested schema](#nestedatt--credit_card))

<a id="nestedatt--credit_card"></a>
### Nested Schema for `credit_card`

Optional:

- `card_type` (String) The credit card type. Default: 'VISA'.
- `cvv` (Number) The credit card CVV.
- `date` (String) The credit card expiration date.
- `number` (Number) The credit card number.
- `postal_code` (String) The postal code attached to the credit card.

</details>
