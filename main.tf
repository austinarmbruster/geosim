terraform {
  required_version = ">= 1.0.0"

  required_providers {
    ec = {
      source = "elastic/ec"
    }

    elasticstack = {
      source = "elastic/elasticstack"
    }
  }
}

provider "ec" {
}

resource "ec_deployment" "user" {
  name = "${var.name}-user"

  region                 = var.region
  version                = var.user_version
  deployment_template_id = var.deployment_template_id


  elasticsearch = {
    hot = {
      size = "8g"
      autoscaling = { }
    }

    config = {
      user_settings_yaml = <<EOF
xpack.security.audit.enabled: true
xpack.security.audit.logfile.events.include: "access_granted,authentication_success"
xpack.security.audit.logfile.events.emit_request_body: true
xpack.security.audit.logfile.events.ignore_filters.example1.users: ["kibana_system"]
    EOF

    }
  }

  kibana = {
    config = {
      user_settings_yaml = <<EOF
server.defaultRoute: "/app/management/insightsAndAlerting/triggersActions/rules"
EOF
    }
  }

  observability = {
    deployment_id = ec_deployment.olly.id
  }

  depends_on = [ ec_deployment.olly ]
}

resource "ec_deployment" "olly" {
  name = "${var.name}-olly"

  region                 = var.region
  version                = var.olly_version
  deployment_template_id = var.deployment_template_id

  elasticsearch = {
    hot = {
      size = "8g"
      autoscaling = { }
    }
  }

  kibana = {
    config = {
      user_settings_yaml = <<EOF
server.defaultRoute: "/app/monitoring"
EOF
    }
  }
}

provider "elasticstack" {
  elasticsearch {
    username  = ec_deployment.olly.elasticsearch_username
    password  = ec_deployment.olly.elasticsearch_password
    endpoints = [ec_deployment.olly.elasticsearch[0].https_endpoint]
  }
}
