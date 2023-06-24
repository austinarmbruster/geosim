output "user-elasticsearch_endpoint" {
  value = ec_deployment.user.elasticsearch.https_endpoint
}

output "user-kibana_endpoint" {
  value = ec_deployment.user.kibana.https_endpoint
}

output "user-elastic_username" {
  value = ec_deployment.user.elasticsearch_username
}

output "user-elastic_password" {
  value = ec_deployment.user.elasticsearch_password
  sensitive = true
}

output "olly-elastic_username" {
  value = ec_deployment.olly.elasticsearch_username
}

output "olly-elastic_password" {
  value = ec_deployment.olly.elasticsearch_password
  sensitive = true
}

output "olly-elasticsearch_endpoint" {
  value = ec_deployment.olly.elasticsearch.https_endpoint
}

output "olly-kibana_endpoint" {
  value = ec_deployment.olly.kibana.https_endpoint
}

output "user-version" {
  value = var.user_version
}

output "olly-version" {
  value = var.olly_version
}