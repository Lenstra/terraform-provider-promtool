terraform {
  required_providers {
    promtool = {
      source = "kevineor/promtool"
    }
  }
}

data "local_file" "prom_file_local" {
  filename = "../rules.yml"
}

resource "local_file" "prom_file_distant" {
  content  = file("../rules.yml")
  filename = "./generated/rules.yml"

  lifecycle {
    precondition {
      condition     = provider::promtool::check_rules(data.local_file.prom_file_local.content)
      error_message = "Rules are not valid"
    }
  }

}

output "name" {
  value = provider::promtool::check_rules(data.local_file.prom_file_local.content)

}
