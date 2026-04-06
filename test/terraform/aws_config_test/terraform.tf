terraform {
  required_version = ">= 1.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.0"
    }
  }

  backend "s3" {
    bucket       = "terraform-state-07027b6d-e4ba-4f0a-abcf-1520f93ebd4d"
    key          = "driftctl-test-infra/terraform.tfstate"
    region       = "us-west-1"
    use_lockfile = true
    encrypt      = true
    # Provide AWS_PROFILE as env var, or this fails
  }
}
