terraform {
  required_providers {
    voipms = {
      source = "vetal-ca-org/voipms"
    }
  }
}

provider "voipms" {
  # username = "you@example.com"
  # password = "api-password-from-portal"
  # api_url  = "https://voip.ms/api/v1/rest.php"
}
