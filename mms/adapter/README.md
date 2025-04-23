# Create Dev Instance
```
aws ec2 run-instances \
  --image-id ami-05f4737ce5776c55b \
  --instance-type g6e.xlarge \
  --key-name GNG_Mac \
  --associate-public-ip-address \
  --security-group-ids sg-0207014882af476de \
  --subnet-id subnet-57a4572f \
  --tag-specifications 'ResourceType=instance,Tags=[{Key=Name,Value=Dataset-Train-dev}]'
 # --iam-instance-profile Name=YourIAMRole
 # --block-device-mappings 'DeviceName=/dev/sda1,Ebs={VolumeSize=100}' \
 

```
{
"Groups": [],
"Instances": [
{
"AmiLaunchIndex": 0,
"ImageId": "ami-05f4737ce5776c55b",
"InstanceId": "i-0dfbdae13535c9d24",
"InstanceType": "g6e.xlarge",
"KeyName": "GNG_Mac",
"LaunchTime": "2025-04-23T17:11:52.000Z",
"Monitoring": {
"State": "disabled"
},
"Placement": {
"AvailabilityZone": "us-west-2a",
"GroupName": "",
"Tenancy": "default"
},
"PrivateDnsName": "ip-172-31-26-35.us-west-2.compute.internal",
"PrivateIpAddress": "172.31.26.35",
"ProductCodes": [],
"PublicDnsName": "",
"State": {
"Code": 0,
"Name": "pending"
},
"StateTransitionReason": "",
"SubnetId": "subnet-57a4572f",
"VpcId": "vpc-4371173b",
"Architecture": "x86_64",
"BlockDeviceMappings": [],
"ClientToken": "937920d1-f3c8-4456-a801-a5fffc361de5",
"EbsOptimized": false,
"EnaSupport": true,
"Hypervisor": "xen",
"NetworkInterfaces": [
{
"Attachment": {
"AttachTime": "2025-04-23T17:11:52.000Z",
"AttachmentId": "eni-attach-0b232025b1daa26d7",
"DeleteOnTermination": true,
"DeviceIndex": 0,
"Status": "attaching",
"NetworkCardIndex": 0
},
"Description": "",
"Groups": [
{
"GroupName": "launch-wizard-5",
"GroupId": "sg-0207014882af476de"
}
],
"Ipv6Addresses": [],
"MacAddress": "02:20:e2:4e:cb:5f",
"NetworkInterfaceId": "eni-07cced74ae0b4e316",
"OwnerId": "078432969830",
"PrivateDnsName": "ip-172-31-26-35.us-west-2.compute.internal",
"PrivateIpAddress": "172.31.26.35",
"PrivateIpAddresses": [
{
"Primary": true,
"PrivateDnsName": "ip-172-31-26-35.us-west-2.compute.internal",
"PrivateIpAddress": "172.31.26.35"
}
],
"SourceDestCheck": true,
"Status": "in-use",
"SubnetId": "subnet-57a4572f",
"VpcId": "vpc-4371173b",
"InterfaceType": "interface"
}
],
"RootDeviceName": "/dev/sda1",
"RootDeviceType": "ebs",
"SecurityGroups": [
{
"GroupName": "launch-wizard-5",
"GroupId": "sg-0207014882af476de"
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
"Value": "Dataset-Train-dev"
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
"ReservationId": "r-0338aeed5d925198f"
}