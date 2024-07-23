resource "local_file" "prom_config_file_distant" {
  content  = local.config
  filename = "./generated/config.yml"

  lifecycle {
    precondition {
      condition     = local.is_config_valid
      error_message = "Rules are not valid"
    }
  }
}

locals {
  config = <<-EOT
global:
  scrape_interval:     15s
  evaluation_interval: 15s

alerting:
  alertmanagers:
  - static_configs:
    - targets:
      - localhost:9093

scrape_configs:
  - job_name: 'prometheus'
    static_configs:
    - targets: ['localhost:9090']

  - job_name: 'node_exporter'
    static_configs:
    - targets: ['localhost:9100']

  - job_name: 'alertmanager'
    static_configs:
    - targets: ['localhost:9093']
    EOT

  is_config_valid = provider::promtool::check_config(local.config)
}

output "config" {
  value = local.is_config_valid

}
