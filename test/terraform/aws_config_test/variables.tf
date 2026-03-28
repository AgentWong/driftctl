variable "aws_profile" {
  description = "AWS CLI profile to use"
  type        = string
}

variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "us-west-2"
}
