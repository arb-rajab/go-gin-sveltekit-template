output "namespace" {
  description = "Namespace the workload should be applied into (pass it to ../k8s/rollout.sh)."
  value       = kubernetes_namespace.app.metadata[0].name
}

output "app_secret_name" {
  description = "Name of the Kubernetes Secret the Deployments reference via envFrom."
  value       = kubernetes_secret.app.metadata[0].name
}

output "cluster_issuer_name" {
  description = "cert-manager ClusterIssuer name referenced by ../k8s/base/ingress.yaml's cert-manager.io/cluster-issuer annotation."
  value       = "starterapp-letsencrypt"
}

output "acme_server_in_use" {
  description = "Which ACME server is configured. Defaults to Let's Encrypt staging (untrusted certificates) until use_staging_acme=false is set deliberately."
  value       = local.acme_server
}
