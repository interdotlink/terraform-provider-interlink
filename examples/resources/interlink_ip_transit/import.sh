# Import an existing IP Transit service by its numeric service ID.
# The service's port already exists, so an imported service is always
# reconstructed with an `existing_port` block (never `new_port`); write your
# configuration to use `existing_port` before importing. The write-only fields
# `bgpsession_password` and `sync_from_pdb`, and the optional fields
# `outbound_advertisement`, `aggregated_billing`, and `purchase_reference`, are
# not read back and are left unset on import.
terraform import interlink_ip_transit.example 11
