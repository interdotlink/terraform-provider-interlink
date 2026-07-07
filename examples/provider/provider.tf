terraform {
  required_providers {
    interlink = {
      source = "interdotlink/interlink"
    }
  }
}

provider "interlink" {
  # API key created under "API Keys" in the Inter.link portal.
  # Prefer supplying it via a variable or TF_VAR_interlink_api_key
  # rather than hard-coding it here.
  api_key = var.interlink_api_key
}
