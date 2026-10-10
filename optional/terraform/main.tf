# Platform layer. Owns cluster-level infrastructure only: the namespace,
# application secrets, the ingress controller, and cert-manager's Let's Encrypt
# issuer. The application workload (Deployments, Services, the migration Job)
# lives in ../k8s and is applied separately by ../k8s/rollout.sh, because the
# platform changes rarely and the workload changes on every release.

resource "kubernetes_namespace" "app" {
  metadata {
    name = var.namespace
    labels = {
      "app.kubernetes.io/part-of"    = "starterapp"
      "app.kubernetes.io/managed-by" = "terraform"
    }
  }
}

# Mirrors .env.example. DATABASE_URL is assembled once, from the same inputs
# Postgres itself is configured with: the backend and the Postgres StatefulSet
# must agree on these credentials, which is why both read from this one Secret.
resource "kubernetes_secret" "app" {
  metadata {
    name      = "starterapp-app-secrets"
    namespace = kubernetes_namespace.app.metadata[0].name
    labels = {
      "app.kubernetes.io/part-of" = "starterapp"
    }
  }

  data = {
    POSTGRES_USER     = var.postgres_user
    POSTGRES_PASSWORD = var.postgres_password
    POSTGRES_DB       = var.postgres_db
    DATABASE_URL = format(
      "postgres://%s:%s@postgres:5432/%s?sslmode=disable",
      var.postgres_user,
      var.postgres_password,
      var.postgres_db,
    )
    SESSION_SIGNING_SECRET = var.session_signing_secret
  }

  type = "Opaque"
}

# --- Ingress controller ---
# Terminates HTTPS for the browser-facing frontend. cert-manager (below) gets it
# a publicly trusted certificate.
resource "helm_release" "ingress_nginx" {
  name             = "ingress-nginx"
  repository       = "https://kubernetes.github.io/ingress-nginx"
  chart            = "ingress-nginx"
  version          = "4.11.3"
  namespace        = "ingress-nginx"
  create_namespace = true

  set {
    name  = "controller.publishService.enabled"
    value = "true"
  }
}

# --- cert-manager ---
# Automates Let's Encrypt issuance and renewal for var.ingress_hostname.
resource "helm_release" "cert_manager" {
  name             = "cert-manager"
  repository       = "https://charts.jetstack.io"
  chart            = "cert-manager"
  version          = "v1.15.3"
  namespace        = "cert-manager"
  create_namespace = true

  set {
    name  = "crds.enabled"
    value = "true"
  }
}

locals {
  acme_server = var.use_staging_acme ? "https://acme-staging-v02.api.letsencrypt.org/directory" : "https://acme-v02.api.letsencrypt.org/directory"
}

# A ClusterIssuer rather than a namespaced Issuer, so a second namespace (for
# example a staging copy) can use it without re-creating it.
resource "kubernetes_manifest" "letsencrypt_issuer" {
  manifest = {
    apiVersion = "cert-manager.io/v1"
    kind       = "ClusterIssuer"
    metadata = {
      name = "starterapp-letsencrypt"
    }
    spec = {
      acme = {
        server = local.acme_server
        email  = var.acme_email
        privateKeySecretRef = {
          name = "starterapp-letsencrypt-account-key"
        }
        solvers = [
          {
            http01 = {
              ingress = {
                ingressClassName = "nginx"
              }
            }
          }
        ]
      }
    }
  }

  depends_on = [helm_release.cert_manager]
}
