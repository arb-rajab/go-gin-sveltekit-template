variable "kubeconfig_path" {
  description = "Path to the kubeconfig for the cluster you already control. This module never provisions the cluster itself."
  type        = string
  default     = "~/.kube/config"
}

variable "kube_context" {
  description = "kubeconfig context to use. Empty string uses the current context."
  type        = string
  default     = ""
}

variable "namespace" {
  description = "Kubernetes namespace the platform resources and workload live in."
  type        = string
  default     = "starterapp"
}

variable "acme_email" {
  description = "Email address cert-manager's ClusterIssuer registers with Let's Encrypt for expiry and revocation notices."
  type        = string
}

variable "ingress_hostname" {
  description = "Public DNS hostname the app will be reachable at (for example starterapp.example.com). Must already point at the ingress controller's external IP or load balancer; this module does not manage DNS."
  type        = string
}

variable "use_staging_acme" {
  description = "Use Let's Encrypt's staging ACME server (untrusted certificates, no rate limits) instead of production. Defaults to true so a first apply cannot burn a production rate-limit slot; flip to false once the hostname and DNS work end to end."
  type        = bool
  default     = true
}

# --- Application secrets ---
# These mirror .env.example so the Kubernetes target needs the same inputs
# Compose does.

variable "postgres_user" {
  description = "Postgres role the app connects as."
  type        = string
  default     = "starterapp"
}

variable "postgres_password" {
  description = "Postgres password. No default: a production secret must be supplied explicitly."
  type        = string
  sensitive   = true
}

variable "postgres_db" {
  description = "Postgres database name."
  type        = string
  default     = "starterapp"
}

variable "session_signing_secret" {
  description = "Base64-encoded HMAC signing key (decodes to at least 32 bytes) for the session cookie. Required: the backend refuses to start without it. Generate with: openssl rand -base64 32"
  type        = string
  sensitive   = true

  validation {
    condition     = length(var.session_signing_secret) > 0
    error_message = "session_signing_secret is required. Generate one with: openssl rand -base64 32"
  }
}

# Container image tags are deliberately not Terraform inputs: per-release image
# versions belong to deploy/k8s/overlays/production/kustomization.yaml (set by
# rollout.sh), not to this module's plan/apply cycle. Terraform owns platform
# state that changes rarely; the image tag changes on every deploy.
