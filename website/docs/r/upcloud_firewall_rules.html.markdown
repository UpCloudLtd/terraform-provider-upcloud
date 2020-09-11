---
layout: "upcloud"
page_title: "UpCloud: resource_upcloud_firewall_rules"
sidebar_current: "docs-upcloud-resource-upcloud-firewall-rules"
description: |-
  Provides a list UpCloud firewall rules
---

# upcloud_firewall_rules

This resource represents a generated list of UpCloud firewall rules.  Firewall rules are used in conjunction with UpCloud servers. 
Each server has its own firewall rules. The firewall is enabled on all network interfaces except ones attached to private virtual networks.
The maximum number of firewall rules per server is 1000.

## Example Usage

The following example defines a server and then links the server to a single firewall rule.
The list of firewall rules applied to the server can be expanded by providing additional `server_firewall_rules` blocks.

```hcl
  resource "upcloud_server" "my_server" {
    zone     = "fi-hel1"
    hostname = "debian.example.com"
    plan     = "1xCPU-2GB"
    firewall = true
  
    storage_devices {
      action = "create"
      size   = 10
      tier   = "maxiops"
    }
  
    network_interface {
      type = "utility"
    }
  
  }
  
  resource "upcloud_firewall_rules" "my_server" {
  
    server_id = upcloud_server.my_server.id
  
    firewall_rule {
  
      action = "accept"
      comment = "Allow SSH from this network"
      destination_port_end = "22"
      destination_port_start = "22"
      direction = "in"
      family = "IPv4"
      protocol = "tcp"
      source_address_end = "192.168.1.255"
      source_address_start = "192.168.1.1"
    }
  
  }
```

## Argument Reference

The following arguments are supported:

* `server_id` - (Required) The unique id of the server to be protected the firewall rules.

* `firewall_rule` - (Required) A customisable list of firewall rules that are to be attached to the server. See [Firewall Rules](#firewall-rules) below for more details.


### Firewall Rules

Each of the `firewall_rule` blocks represent a single firewall rule that is the be attached to the server specified through `server_id`.
The order of the `firewall_rule` blocks is used as the order to attach each rule to the server.  Moving the block will change the order they are applied in.

Each `firewall_rule` block supports the following:

* `action` - (Required) Action to take if the rule conditions are met. Accepted value `accept` or `drop`.

* `comment` - (Optional) Freeform comment string for the rule.  Accepted length 0-250 characters.

* `destination_address_end` - (Optional) The destination address range ends from this address.  Required if using `destination_address_start`. 

* `destination_address_start` - (Optional) The destination address range starts from this address.  Required if using `destination_address_end`.

* `destination_port_end` - (Optional) The destination port range ends from this port number. Required if using `destination_port_start`.  Accepted value 1-65535.

* `destination_port_start` - (Optional) The destination port range starts from this port number. Required if using `destination_port_end`. Accepted valye 1-65535.

* `direction` - (Required) - The direction of network traffic this rule will be applied to.

* `family` - (Required) - The address family of new firewall rule, Accepted value `IPv4` or `IPv6`.

* `icmp_type` - (Optional) - The ICMP type.  Accepted value 0-255.

* `protocol` - (Optional) - The protocol this rule will be applied to.  Accepted values `tcp`, `udp` or `icmp`.

* `source_address_end` - (Optional) The source address range ends from this address. Required if using `source_address_start`

* `source_address_start` - (Optional) The source address range starts from this address. Required if using `source_address_end`

* `source_port_end` - (Optional) The source port range ends from this port number.  Required if using `source_port_start`.  Accepted value 1-65535.

* `source_port_start` - (Optional) The source port range starts from this port number.  Required if using `source_port_end`.  Accepted value 1-65535.

## Import

Existing UpCloud firewall rules can be imported into the current Terraform state through the server id UUID.

```hcl
terraform import upcloud_firewall_rules.my_example_rules 049d7ca2-757e-4fb1-a833-f87ee056547a
```