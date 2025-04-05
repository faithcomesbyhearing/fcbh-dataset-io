### Procedure to set up dbp_timestamp server

# Describe current Dataset-V2
```
aws ec2 describe-instances \
  --instance-ids i-0b22222aa0f43d1a5 \
  --query 'Reservations[0].Instances[0].{SubnetId:SubnetId,SecurityGroups:SecurityGroups}'
  ```

# Get Sample Policy File
```
aws s3api get-bucket-policy --bucket dataset-io --query Policy --output text > dataset-io.json
```

# Create AMI
```
aws ec2 create-image \
--instance-id i-0b22222aa0f43d1a5 \
--name "Dataset-V2-AMI" \
--description "AMI created from Dataset-V2 instance"
```

# Create Instance
```
aws ec2 run-instances \
  --image-id ami-xxxxxxxxx \
  --instance-type g6e.xlarge \
  --key-name GNG_Mac \
  --security-group-ids sg-xxxxxxxxx \
  --subnet-id subnet-xxxxxxxxx \
  --iam-instance-profile Name=YourIAMRole
 # --block-device-mappings 'DeviceName=/dev/sda1,Ebs={VolumeSize=100}' \
 # --tag-specifications 'ResourceType=instance,Tags=[{Key=Name,Value=Dataset-V2-Instance}]'
```

# Edit .bash_profile


# Create queue bucket
```
aws s3api create-bucket --bucket timestamp-queue --region us-west-2
```

# Add Permissions
```
aws s3api put-bucket-policy \
  --bucket timestamp-queue \
  --policy file://timestamp-queue.json
  ```

# Create io bucket
```
aws s3api create-bucket --bucket timestamp-io --region us-west-2
```

# Add Permissions
```
aws s3api put-bucket-policy \
  --bucket timestamp-io \
  --policy file://timestamp-io.json
  ```