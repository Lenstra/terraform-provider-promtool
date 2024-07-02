terraform {
  required_providers {
    promtool = {
      source = "lenstra/promtool"
    }
  }
}

data "local_file" "prom_file_local" {
  filename = "../rules.yml"
}

resource "local_file" "prom_file_distant" {
  content  = local.config
  filename = "./generated/rules.yml"

  lifecycle {
    precondition {
      condition     = local.is_config_valid
      error_message = "Rules are not valid"
    }
  }
}

locals {
  config = <<-EOT
    groups:
    - name: example
      rules:
      - alert: HighRequestLatency
        expr: job:request_latency_seconds:mean5m{job="myjob"} > 0.5
        for: 10m
        labels:
          severity: page
        annotations:
          summary: High request latency
    EOT

  is_config_valid = provider::promtool::check_rules(local.config)
}

output "bla" {
  value = local.is_config_valid

}
