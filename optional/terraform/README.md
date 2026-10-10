# Optional module: Terraform platform layer

Terraform for the cluster-level pieces `../k8s` expects: the namespace, the
`starterapp-app-secrets` Secret, an ingress-nginx controller, cert-manager, and
a Let's Encrypt ClusterIssuer. It targets a Kubernetes cluster you already
control; it does not create one. Delete this directory if you manage the cluster
another way.

Nothing here has been applied to a live cluster by the template authors.
`terraform fmt -check` runs in CI; `terraform validate` and a real plan/apply
are yours to run.

## Use

```
cd optional/terraform
cp terraform.tfvars.example terraform.tfvars   # fill in real values; never commit it
terraform init
terraform plan
terraform apply
```

- `use_staging_acme` defaults to `true`, so a first apply issues untrusted
  staging certificates. Set it to `false` once DNS and the ingress work.
- Secrets are stored in Terraform state. Use an encrypted remote backend and
  restrict access to it.
- After applying, edit the hostname patch in
  `../k8s/overlays/production/kustomization.yaml` to match `ingress_hostname`,
  then run `../k8s/rollout.sh`.
