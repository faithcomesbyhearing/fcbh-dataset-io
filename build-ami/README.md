# fcbh-dataset-io infrastructure 

## terraform
A separate GitHub repo *fcbh-infrastructure-live*, provides the following AWS resources:
- storage of secrets such as the Biblebrain API key and OpenAI API key
- creation of S3 buckets for 
    - queue
    - input
    - io
- IAM roles for:
    - access to S3 buckets
    - EC2 Instance permissions
- Security Group to access the EC2 instance
- Launch Template with appropriate sizing, network configuration and instance profile
- Eventbridge Event Bus and trigger (future)
- Codebuild job to produce artifacts based on updates to this repository (future)

## build process (as of March 31)
1. in fcbh-dataset-io, create the AMI
``` bash
cd build-ami
packer build -var-file="env_vars_dev.hcl" -var-file="secrets-dev.hcl" -debug template.pkr.hcl 
# the packer invocation takes about 45 minutes to complete
# when done, packer will output similar to this:
#
#==> Builds finished. The artifacts of successful builds are:
#--> amazon-ebs.my-ami: AMIs were created:
#us-west-2: ami-0ef07bd570a904d5b
```
The AMI is created with a name like "deep-learning-golden-ami-*". This will be used by terraform to find the correct AMI for insertion into the launch template

2. Manually test the AMI
    - launch an EC2 instance from the AMI listed above
    - log in to the instance and verify it is configured as expected
    - in the future, we can add tests to packer so that the AMI will only be created if the tests pass

3. in fcbh-infrastructure-live:
    - directory: /content-creation/dbp-dev-account/us-west-2/dev/deep-learning
    - run terragrunt apply to update the EC2 Launch Template

## to run:
-- start an EC2 instance referencing the launch template
```
ec2 start-instance 
```

