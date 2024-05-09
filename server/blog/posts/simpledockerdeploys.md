# Simple Docker Deploys
I'd like to take some time here to describe how I deploy my personal website, `andrewwillette.com`. For a non-production application maintained solely by myself, I think it's a great solution. All the technologies I use are standard pieces of a modern cloud stack. I've found that maintaining my website is a nice exercise in keeping up-to-date with popular pieces of the cloud pipeline.

## EC2 Instance
My website runs on an [EC2 instance](https://aws.amazon.com/ec2/) in AWS. I use [packer CLI](https://www.packer.io/) to build an AMI with docker installed and running. Below is the [hcl2 script](https://developer.hashicorp.com/packer/guides/hcl) for my website.

```
packer {
  required_plugins {
    amazon = {
      version = ">= 1.2.6"
      source  = "github.com/hashicorp/amazon"
    }
  }
}

source "amazon-ebs" "ubuntu" {
  ami_name      = "ubuntu-docker-{{timestamp}}"
  instance_type = "t3.small"
  region        = "us-east-2"

  source_ami_filter {
    filters = {
      "virtualization-type" = "hvm"
      "root-device-type"    = "ebs"
      # just copy the latest public AMI from
      # searching AMI's in console, don't buy anything
      name                  = "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-20230420"
    }
    owners      = ["<aws_user_id>"]
    most_recent = true
  }

  ssh_username = "ubuntu"
}

build {
  sources = ["source.amazon-ebs.ubuntu"]
  provisioner "shell" {
    # executes everything as sudo
    execute_command = "echo 'packer' | sudo -S env {{ .Vars }} {{ .Path }}"
    # mkdir call is for docker volume caching SSL certs across docker builds
    inline = [
      <<-EOT
        #!/bin/sh
        apt-get update
        apt-get install -y docker.io
        mkdir -p /var/www/.cache
        systemctl enable docker
        systemctl start docker
      EOT
    ]
  }
}
```

With the above script saved as `alpine-docker.pkr.hcl`, executing `packer build alpine-docker.pkr.hcl` outputs an `ami-ID` that I use in the below [terraform](https://www.terraform.io/) script.

```
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.0"
    }
  }
}

provider "aws" {
  region = "us-east-2"
}

resource "aws_key_pair" "willette_key" {
  key_name   = "willette-key"
  public_key = "ssh-rsa <public_key_from_local_machine> andrewwillette@andrewmacbook.local"
}

resource "aws_instance" "willette_website" {
  ami           = "<ami_id_from_packer_output>"
  instance_type = "t3.small"
  tags = {
    Name = "AndrewWilletteDotCom"
  }
  key_name               = aws_key_pair.willette_key.key_name
  vpc_security_group_ids = [aws_security_group.main.id]
  root_block_device {
    volume_size = 30
    volume_type = "gp2"
  }
}

resource "aws_security_group" "main" {
  egress = [
    {
      cidr_blocks      = ["0.0.0.0/0", ]
      description      = ""
      from_port        = 0
      ipv6_cidr_blocks = []
      prefix_list_ids  = []
      protocol         = "-1"
      security_groups  = []
      self             = false
      to_port          = 0
    }
  ]
  ingress = [
    {
      cidr_blocks      = ["0.0.0.0/0", ]
      description      = "SSH ingress"
      from_port        = 22
      ipv6_cidr_blocks = []
      prefix_list_ids  = []
      protocol         = "tcp"
      security_groups  = []
      self             = false
      to_port          = 22
    },
    {
      description      = "http ingress"
      self             = false
      from_port        = 80
      to_port          = 80
      protocol         = "tcp"
      prefix_list_ids  = []
      security_groups  = []
      cidr_blocks      = ["0.0.0.0/0"]
      ipv6_cidr_blocks = []
    },
    {
      description      = "https ingress"
      self             = false
      from_port        = 443
      to_port          = 443
      protocol         = "tcp"
      prefix_list_ids  = []
      security_groups  = []
      cidr_blocks      = ["0.0.0.0/0"]
      ipv6_cidr_blocks = []
    }
  ]
}
```

The terraform script also includes details for an ssh key. This is a public-key associated with a private-key on my local machine; this SSH connection comes into play later. Port ingress/egress rules are also declared for ssh, http, and https.

Executing `terraform plan` and `terraform apply` with the above script defined in the current directory as `website.tf` will deploy the EC2 instance into AWS.

## Docker over SSH

I can now configure docker builds on my local machine to execute builds on the recently-deployed EC2 instance. A [docker context](https://docs.docker.com/engine/context/working-with-contexts/) on my personal machine connects to my EC2 instance via SSH using the command `docker context create --docker host=ssh://ubuntu@<aws_public_ip> personalwebsite`. This is utilizing the ssh-key I configured in the above terraform script.

I now have a [single bash script I execute locally](https://github.com/andrewwillette/andrewwillette.com/blob/main/deploy-prod.sh) which will build the website's docker container from my local machine's code and deploy it to AWS!
