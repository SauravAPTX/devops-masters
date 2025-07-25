aws_region   = "ap-south-1"
project_name = "devops-masters"
github_repo  = "https://github.com/SauravAPTX/devops-masters.git"

environment  = "production"
# ecr_registry = "237206024543.dkr.ecr.ap-south-1.amazonaws.com/devops-app"

vpc_id = "vpc-0c9dbce2ea7089119"

subnet_ids = [
  "subnet-0596048fb174d0baa",
  "subnet-09a77d5d24863c1f9",
  "subnet-0320761b5120b7ee3"
]