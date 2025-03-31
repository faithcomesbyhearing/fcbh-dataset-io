variable "region" {
  type    = string
  default = "us-west-2"
}

variable "aws_profile" {
  type    = string
  default = "dbp-dev"
}
variable "vpc_id" {
  type    = string
  default = "vpc-0b6a6785e74d18db3"
}

variable "subnet_id" {
  type    = string
  default = "subnet-086d451813d884dd0"
}
variable "ami_name_prefix" {
  type    = string
  default = "deep-learning-golden-ami"
}

variable "instance_type" {
  type    = string
  default = "g6e.xlarge"
}

variable "source_ami" {
  type    = string
  default = "ami-013e597d66c833276"
}
variable "ssh_username" {
  type    = string
  default = "ubuntu"
}

variable "env_var_filename" {
  type    = string
  default = "env_vars.txt"
}

# these must be provided during packer invocation
variable "fcbh_dataset_queue" {
  type    = string
}
variable "fcbh_dataset_io_bucket" {
  type    = string
}
variable "openai_api_key" {
  type    = string
}
variable "fcbh_dbp_key" {
  type    = string
}


## should be unchanged below here

packer {
  required_plugins {
    amazon = {
      version = ">= 1.0.0"
      source  = "github.com/hashicorp/amazon"
    }
  }
}

# Note: The AMI is based on ubuntu 22.04
source "amazon-ebs" "my-ami" {
  ami_name      = "${var.ami_name_prefix}-{{timestamp}}" 
  instance_type = var.instance_type 
  region        = var.region
  source_ami    = var.source_ami  
  ssh_username  = var.ssh_username

  ami_block_device_mappings {
    device_name = "/dev/sda1"
    volume_size = 128
    volume_type = "gp3"
    delete_on_termination = true
  }
  

  # note: user-data.sh will be provided by the launch template (must faster)

  profile = var.aws_profile

  vpc_id           = var.vpc_id
  subnet_id        = var.subnet_id

  ami_description = "Custom deep-learning AMI (ubuntu 22.04) built with Packer"
  tags = {
    Name = "Deep Learning"
  }
}

build {
  sources = [
    "source.amazon-ebs.my-ami"
  ]


  # set config files and environment variables that are relative to another environment variable
  provisioner "file" {
    source      = "cloudwatch-nvidia.json"  
    destination = "/tmp/cloudwatch-nvidia.json"
  }
  provisioner "shell" {
    inline = [
      "sudo mv /tmp/cloudwatch-nvidia.json  /opt/aws/amazon-cloudwatch-agent/etc/amazon-cloudwatch-agent.json",
      "sudo chmod 644 /opt/aws/amazon-cloudwatch-agent/etc/amazon-cloudwatch-agent.json",
      "sudo echo export HOME=/home/ubuntu >> /home/ubuntu/.bash_profile",
      "sudo echo export GOPATH=$HOME/go  >> /home/ubuntu/.bash_profile",
      "sudo echo export GOPROJ=$HOME/go/src  >> /home/ubuntu/.bash_profile",
      "sudo echo export PATH=$PATH:$GOPATH/bin  >> /home/ubuntu/.bash_profile",
      "sudo echo export FCBH_DATASET_DB=$HOME/data  >> /home/ubuntu/.bash_profile",
      "sudo echo export FCBH_DATASET_FILES=$HOME/data/download  >> /home/ubuntu/.bash_profile",
      "sudo echo export FCBH_DATASET_TMP=$HOME/data/tmp  >> /home/ubuntu/.bash_profile",
      "sudo echo export FCBH_DATASET_LOG_FILE=$HOME/dataset.log  >> /home/ubuntu/.bash_profile"
    ]
  }

  # set common environment variables into .bash_profile
  provisioner "file" {
    source      = "bash-vars.txt"  
    destination = "/tmp/env_vars.txt"
  }
  provisioner "shell" {
      inline = [
        "echo 'Setting environment variables from file:'",
        "while IFS='=' read -r var val; do",
        "  echo \"Setting $var to $val\"",
        "  echo \"export $var='$val'\" >> /home/ubuntu/.bash_profile",
        "done < /tmp/env_vars.txt"
      ]
  }

  # set environment-specific variables (eg the S3 buckets are different from dev to prod)
  provisioner "shell" {
      environment_vars = [
        "FCBH_DATASET_QUEUE=${var.fcbh_dataset_queue}",
        "FCBH_DATASET_IO_BUCKET=${var.fcbh_dataset_io_bucket}",
      ]
      inline = [
        "echo 'export FCBH_DATASET_QUEUE='$FCBH_DATASET_QUEUE >> /home/ubuntu/.bash_profile",
        "echo 'export FCBH_DATASET_IO_BUCKET='$FCBH_DATASET_IO_BUCKET >> /home/ubuntu/.bash_profile",
      ]
  }

  # set secrets
  provisioner "shell" {
      environment_vars = [
        "OPENAI_API_KEY=${var.openai_api_key}",
        "FCBH_DBP_KEY=${var.fcbh_dbp_key}",
      ]
      inline = [
        "echo 'export OPENAI_API_KEY='$OPENAI_API_KEY >> /home/ubuntu/.bash_profile",
        "echo 'export FCBH_DBP_KEY='$FCBH_DBP_KEY >> /home/ubuntu/.bash_profile",
      ]
  }
  # install code
  provisioner "shell" {
    script = "provision-ami.sh"
  }
}