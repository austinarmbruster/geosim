variable "name" {
    type = string
    default = "tracking-containment"
}
variable "user_version" {
  type = string
  default = "8.8.1"
}

variable "olly_version" {
  type = string
  default = "8.8.1"
}

variable "region" {
  type = string
  default = "us-east-1"
}

variable "deployment_template_id" {
  type = string
  default = "aws-io-optimized-v2"
}

