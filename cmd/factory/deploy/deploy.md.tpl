# Deploys infrastructure changes (e.g., terraform apply, pulumi up).

Detects the project type (Terraform, Pulumi), explains the deployment action,
and executes the deployment after confirmation.

- Terraform: Requires 'tfplan.out' from 'contextvibes factory plan'. Runs 'terraform apply tfplan.out'.
- Pulumi: Runs 'pulumi up', which internally includes a preview and confirmation.
