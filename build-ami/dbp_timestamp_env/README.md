# Procedure to set up dbp_timestamp server

## Describe current dbp-etl-dev
```
aws ec2 describe-instances \
  --instance-ids i-01613e74dc90be1e3 \
  --query 'Reservations[0].Instances[0].{SubnetId:SubnetId,SecurityGroups:SecurityGroups}'
  ```
{
"SubnetId": "subnet-0514270361bd9075e",
"SecurityGroups": [
{
"GroupName": "dbp-etl",
"GroupId": "sg-02d16f1ea3ccb4c6c"
},
{
"GroupName": "Bastion Security Group-20200323112037932500000001",
"GroupId": "sg-0b3bc5538486de5f5"
}
]
}


## Get user roles
```
aws ec2 describe-instances --instance-ids i-0b22222aa0f43d1a5 --query "Reservations[*].Instances[*].IamInstanceProfile" --output text
```
Nothing returned

## Get Sample Policy File
```
aws s3api get-bucket-policy --bucket dataset-io --profile sandeep --output json | jq -r '.Policy' > dataset-io.json
```

## Create AMI
```
aws ec2 create-image \
--instance-id i-0b22222aa0f43d1a5 \
--name "Dataset-V2-AMI" \
--description "AMI created from Dataset-V2 instance" \
--reboot
```
{
    ##"ImageId": "ami-09287325315caffd5"
    "ImageId": "ami-05f4737ce5776c55b"

}

# Create Instance
```
aws ec2 run-instances \
  --image-id ami-05f4737ce5776c55b \
  --instance-type g6e.xlarge \
  --key-name GNG_Mac \
  --security-group-ids sg-02d16f1ea3ccb4c6c \
  --subnet-id subnet-0514270361bd9075e \
  --tag-specifications 'ResourceType=instance,Tags=[{Key=Name,Value=Dataset-TS-dev}]'
 # --iam-instance-profile Name=YourIAMRole
 # --block-device-mappings 'DeviceName=/dev/sda1,Ebs={VolumeSize=100}' \
```
{
"Groups": [],
"Instances": [
{
"AmiLaunchIndex": 0,
"ImageId": "ami-05f4737ce5776c55b",
"InstanceId": "i-0526bb0caca7826e3",
"InstanceType": "g6e.xlarge",
"KeyName": "GNG_Mac",
"LaunchTime": "2025-04-09T14:32:04.000Z",
"Monitoring": {
"State": "disabled"
},
"Placement": {
"AvailabilityZone": "us-west-2a",
"GroupName": "",
"Tenancy": "default"
},
"PrivateDnsName": "ip-172-17-21-39.us-west-2.compute.internal",
"PrivateIpAddress": "172.17.21.39",
"ProductCodes": [],
"PublicDnsName": "",
"State": {
"Code": 0,
"Name": "pending"
},
"StateTransitionReason": "",
"SubnetId": "subnet-0514270361bd9075e",
"VpcId": "vpc-0b6a6785e74d18db3",
"Architecture": "x86_64",
"BlockDeviceMappings": [],
"ClientToken": "234bbd6e-9de5-45b4-ba51-99395eda431b",
"EbsOptimized": false,
"EnaSupport": true,
"Hypervisor": "xen",
"NetworkInterfaces": [
{
"Attachment": {
"AttachTime": "2025-04-09T14:32:04.000Z",
"AttachmentId": "eni-attach-074edd39fb0f12959",
"DeleteOnTermination": true,
"DeviceIndex": 0,
"Status": "attaching",
"NetworkCardIndex": 0
},
"Description": "",
"Groups": [
{
"GroupName": "dbp-etl",
"GroupId": "sg-02d16f1ea3ccb4c6c"
}
],
"Ipv6Addresses": [],
"MacAddress": "02:ab:f0:8e:23:8d",
"NetworkInterfaceId": "eni-0a7f2fed5ec92897e",
"OwnerId": "078432969830",
"PrivateDnsName": "ip-172-17-21-39.us-west-2.compute.internal",
"PrivateIpAddress": "172.17.21.39",
"PrivateIpAddresses": [
{
"Primary": true,
"PrivateDnsName": "ip-172-17-21-39.us-west-2.compute.internal",
"PrivateIpAddress": "172.17.21.39"
}
],
"SourceDestCheck": true,
"Status": "in-use",
"SubnetId": "subnet-0514270361bd9075e",
"VpcId": "vpc-0b6a6785e74d18db3",
"InterfaceType": "interface"
}
],
"RootDeviceName": "/dev/sda1",
"RootDeviceType": "ebs",
"SecurityGroups": [
{
"GroupName": "dbp-etl",
"GroupId": "sg-02d16f1ea3ccb4c6c"
}
],
"SourceDestCheck": true,
"StateReason": {
"Code": "pending",
"Message": "pending"
},
"Tags": [
{
"Key": "Name",
"Value": "Dataset-TS-dev"
}
],
"VirtualizationType": "hvm",
"CpuOptions": {
"CoreCount": 2,
"ThreadsPerCore": 2
},
"CapacityReservationSpecification": {
"CapacityReservationPreference": "open"
},
"MetadataOptions": {
"State": "pending",
"HttpTokens": "optional",
"HttpPutResponseHopLimit": 1,
"HttpEndpoint": "enabled",
"HttpProtocolIpv6": "disabled",
"InstanceMetadataTags": "disabled"
},
"EnclaveOptions": {
"Enabled": false
},
"BootMode": "uefi-preferred",
"PrivateDnsNameOptions": {
"HostnameType": "ip-name",
"EnableResourceNameDnsARecord": false,
"EnableResourceNameDnsAAAARecord": false
},
"MaintenanceOptions": {
"AutoRecovery": "default"
},
"CurrentInstanceBootMode": "uefi"
}
],
"OwnerId": "078432969830",
"ReservationId": "r-09c3cc8afa92ab13b"
}

# Edit .bash_profile
timestamp-queue
timestamp-io

# Create queue bucket
```
aws s3api create-bucket \
    --bucket timestamp-queue \
    --region us-west-2 \
    --create-bucket-configuration LocationConstraint=us-west-2
```

# Add Permissions
```
aws s3api put-bucket-policy \
  --bucket timestamp-queue \
  --policy file://timestamp-queue.json
  ```
```
aws s3api get-bucket-policy --bucket timestamp-queue --output json 
```

# Create io bucket
```
aws s3api create-bucket \
    --bucket timestamp-io \
    --region us-west-2 \
    --create-bucket-configuration LocationConstraint=us-west-2
```

# Add Permissions
```
aws s3api put-bucket-policy \
  --bucket timestamp-io \
  --policy file://timestamp-io.json
  ```
```
aws s3api get-bucket-policy --bucket timestamp-io --output json 
```