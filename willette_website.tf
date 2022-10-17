variable "sshkey" {
  description = "Public ssh key"
  default     = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQCwDnlIDc2KfSJ/9Cq2jmZ3VuV9iFORuOyQME/Q0rebyPKes0Lx7wxwLW+D5woGRNOxMmckL2/xyIx0RKZkHzGFg6GyKEDftk2D/f298dwHB2UCSpF1hNXW/JFKrMSPNfW1Sa71hgUpVBmB4qYRnYeEM16oohIqe4JWBfcL9HDQxhmyqkBnKfzG4hsxyaxfWwsBA+kxwVlb08Sh++h5XbdJVMpWLw7UQsL5evZZEXFw9xJZqYo+VEBcomaXbF4iLcj7vX6v5xoF08Kx+YJqnvTEVIiLLgMvzoPHSTeH9K+Bf2Fvrn/xetyrYNjlCaA3J9y7WOxZbTjtXfgUHia/a/qr2rei03zBgJNBNPbdTPEZM7HE6V1jUun7LqK3h1a1BWnmtfRNEA3n17rXG3a7Ohvdx3qqwiEIGYoMP9u03nXT4bHCc0XJb5XMEaNKdwyozdbCVHfUsGkATOs+phPN+VrapWPzhhzZzDGHWY6l7LvwV8y61Z5sVGS3vM8FdXgi/fk= andrewwillette@andrewmacbook.local"
}

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
  public_key = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQCwDnlIDc2KfSJ/9Cq2jmZ3VuV9iFORuOyQME/Q0rebyPKes0Lx7wxwLW+D5woGRNOxMmckL2/xyIx0RKZkHzGFg6GyKEDftk2D/f298dwHB2UCSpF1hNXW/JFKrMSPNfW1Sa71hgUpVBmB4qYRnYeEM16oohIqe4JWBfcL9HDQxhmyqkBnKfzG4hsxyaxfWwsBA+kxwVlb08Sh++h5XbdJVMpWLw7UQsL5evZZEXFw9xJZqYo+VEBcomaXbF4iLcj7vX6v5xoF08Kx+YJqnvTEVIiLLgMvzoPHSTeH9K+Bf2Fvrn/xetyrYNjlCaA3J9y7WOxZbTjtXfgUHia/a/qr2rei03zBgJNBNPbdTPEZM7HE6V1jUun7LqK3h1a1BWnmtfRNEA3n17rXG3a7Ohvdx3qqwiEIGYoMP9u03nXT4bHCc0XJb5XMEaNKdwyozdbCVHfUsGkATOs+phPN+VrapWPzhhzZzDGHWY6l7LvwV8y61Z5sVGS3vM8FdXgi/fk= andrewwillette@andrewmacbook.local"
}

resource "aws_instance" "willette_website" {
  ami           = "ami-04347cf5004fed072"
  instance_type = "t3.small"

  tags = {
    Name = "AndrewWilletteDotCom"
  }
  key_name               = aws_key_pair.willette_key.key_name
  vpc_security_group_ids = [aws_security_group.main.id]

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
      description      = "http ingress for frontend"
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
      description      = "http ingress for backend"
      self             = false
      from_port        = 9099
      to_port          = 9099
      protocol         = "tcp"
      prefix_list_ids  = []
      security_groups  = []
      cidr_blocks      = ["0.0.0.0/0"]
      ipv6_cidr_blocks = []
    }
  ]
}
