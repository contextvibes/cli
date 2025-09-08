# Generates an execution plan for infrastructure changes.

Detects the project type (Terraform, Pulumi) and runs the appropriate
command to generate an execution plan, showing expected changes.

- Terraform: Runs 'terraform plan -out=tfplan.out'
- Pulumi: Runs 'pulumi preview'
