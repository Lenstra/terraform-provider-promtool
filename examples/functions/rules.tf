resource "local_file" "prom_rule_file_distant" {
  content  = local.rules_config
  filename = "./generated/rules.yml"

  lifecycle {
    precondition {
      condition     = local.is_rules_config_valid
      error_message = "Rules are not valid"
    }
  }
}

locals {
  rules_config = <<-EOT
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

  is_rules_config_valid = provider::promtool::check_rules(local.rules_config)
}

output "rules" {
  value = local.is_rules_config_valid

}
