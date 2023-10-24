# terraform-provider-gitea

Terraform Gitea Provider

This repo is mirrored from https://gitea.com/gitea/terraform-provider-gitea please send all issues and pull requests there.

## Usage

This is not a 1.0 release, so usage is subject to change!

```terraform
terraform {
  required_providers {
    gitea = {
      source = "go-gitea/gitea"
      version = "0.3.0"
    }
  }
}

provider "gitea" {
  base_url = var.gitea_url # optionally use GITEA_BASE_URL env var
  token    = var.gitea_token # optionally use GITEA_TOKEN env var

  # Username/Password authentication is mutally exclusive with token authentication
  # username = var.username # optionally use GITEA_USERNAME env var
  # password = var.password # optionally use GITEA_PASSWORD env var

  # A file containing the ca certificate to use in case ssl certificate is not from a standard chain
  cacert_file = var.cacert_file 
  
  # If you are running a gitea instance with self signed TLS certificates
  # and you want to disable certificate validation you can deactivate it with this flag
  insecure = false 
}

resource "gitea_repository" "test" {
  username     = "lerentis"
  name         = "test"
  private      = true
  issue_labels = "Default"
  license      = "MIT"
  gitignores   = "Go"
}

resource "gitea_repository" "mirror" {
  username                     = "lerentis"
  name                         = "terraform-provider-gitea-mirror"
  description                  = "Mirror of Terraform Provider"
  mirror                       = true
  migration_clone_addresse     = "https://git.uploadfilter24.eu/lerentis/terraform-provider-gitea.git"
  migration_service            = "gitea"
  migration_service_auth_token = var.gitea_mirror_token
}

resource "gitea_org" "test_org" {
  name = "test-org"
}

resource "gitea_repository" "org_repo" {
  username = gitea_org.test_org.name
  name = "org-test-repo"
}

```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## History

This codebase was created at https://gitea.com/gitea/terraform-provider-gitea, was forked by @lerentis, and then their changes were merged back into the original repo. Thank you to everyone who contributed!
