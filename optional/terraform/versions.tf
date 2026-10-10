# Provider pins for the Kubernetes platform layer.
#
# This module deliberately does not declare a cloud provider (AWS, GCP, Azure,
# Hetzner, ...). It manages resources inside a Kubernetes cluster you already
# control (bring your own cluster). Point `var.kubeconfig_path` and
# `var.kube_context` at whatever cluster that is.

terraform {
  required_version = ">= 1.5.0"

  required_providers {
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 3.3"
    }
    helm = {
      source  = "hashicorp/helm"
      version = "~> 2.15"
    }
  }
}

provider "kubernetes" {
  config_path    = var.kubeconfig_path
  config_context = var.kube_context
}

provider "helm" {
  kubernetes {
    config_path    = var.kubeconfig_path
    config_context = var.kube_context
  }
}
