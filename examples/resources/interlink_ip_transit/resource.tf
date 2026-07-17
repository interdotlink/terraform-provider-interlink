# Manage an IP Transit service delivered on a newly provisioned port.
resource "interlink_ip_transit" "example" {
  bgpsession_asn             = 64500 # your public AS number
  bgpsession_as_set          = "AS-EXAMPLE"
  bgpsession_password        = var.bgp_password # sensitive; write-only
  bgpsession_prefix_limit_v4 = 1000
  bgpsession_prefix_limit_v6 = 100
  prefix_v4_size             = 30  # 30 or 31
  prefix_v6_size             = 126 # 126 or 127
  term                       = 12  # months
  sync_from_pdb              = true

  # Provide exactly one port block: new_port, existing_port, or existing_lag.
  new_port {
    location  = "FRA1-DE"
    bandwidth = 1000 # committed data rate in Mbps
    port_type = "10G-LR (SFP+)"
    vlan_id   = 100
    vlan_type = "Tagged"
  }
}

# Attach to an existing port instead of provisioning a new one:
#
#   existing_port {
#     id        = 42
#     bandwidth = 1000
#     vlan_id   = 100
#     vlan_type = "Tagged"
#   }
#
# ...or to an existing LAG:
#
#   existing_lag {
#     lag_id    = 7
#     bandwidth = 1000
#   }
