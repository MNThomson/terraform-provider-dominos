Terraform Provider for Dominos Pizza
==================
# Quickstart

Then write your config.  Here's a sample config - a variation on this worked for me last night.

```hcl
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

  credit_card {
    number = 123456789101112
    cvv    = 1314
    date   = "15/16"
    zip    = 18192
  }
}

data "dominos_address" "addr" {
  street = "123 Main St"
  city   = "Anytown"
  state  = "WA"
  zip    = "02122"
}

data "dominos_store" "store" {
  address_url_object = data.dominos_address.addr.url_object
}

data "dominos_menu_item" "item" {
  store_id     = data.dominos_store.store.store_id
  query_string = ["philly", "medium"]
}

resource "dominos_order" "order" {
  address_api_object = data.dominos_address.addr.api_object
  item_codes         = ["${data.dominos_menu_item.item.matches.0.code}"]
  store_id           = data.dominos_store.store.store_id
}
```


`terraform init` as usual and `plan`!  `apply` when ready - but use caution, since this is going to charge you money.

## Credit

Massive credit to [nat-henderson](https://github.com/nat-henderson/terraform-provider-dominos): they built the kitchen, assembled the wood fired oven, and perfected the recipe. I am merely the waiter serving this pizza to the masses.
