# Simple Docker Deploys
I'd like to take some time to describe how I build and deploy my personal website, `andrewwillette.com`. The website is a non-critical web application maintained by one person, me. With those "requirements", I think I have a nice solution. It uses a lot of the popular cloud technologies. That makes maintaining it more interesting and rewarding.

## Echo HTTPS Server
The webserver is implemented using the `go` [echo framework](https://github.com/labstack/echo). [Go templating](https://pkg.go.dev/html/template) is used for frontend requirements and as a result the entire website can be deployed as a single go binary; there's no frontend/backend complexity. The echo framework [provides an API](https://echo.labstack.com/docs/cookbook/auto-tls#server) for management of TLS certificates. This is one of the primary reasons I haven't switched to the [stdlib webserver](https://pkg.go.dev/net/http) yet.

```go
func startServer() {
    e := echo.New()
    e.Pre(middleware.HTTPSRedirect())
    e.AutoTLSManager.HostPolicy = autocert.HostWhitelist("andrewwillette.com")
    // getSSLCacheDir return directory for ssl cache
    const sslCacheDir = "/var/www/.cache"
    e.AutoTLSManager.Cache = autocert.DirCache(sslCacheDir)
    // various other routing and middleware omitted
    go func(c *echo.Echo) {
      e.Logger.Fatal(e.Start(":80"))
    }(e)
    e.Logger.Fatal(e.StartAutoTLS(":443"))
}
```

After running into issues where I couldn't access my website post-redeploy, I persisted the docker container's `/var/www/.cache` directory across docker deploys as a [docker volume](https://docs.docker.com/storage/volumes/). If this is not done, the SSL certificate updates each deploy. Clients (browsers) consequentially don't trust the newly-deployed service with its changed certificate. My docker compose file, `docker-compose-prod.yml` is below. It shows how the certificate directory is configured as a container "volume" on the host-machine.

```
version: '3'

services:
  andrewwillette:
    build:
      context: .
    image: andrewwillette-dot-com:1.0
    environment:
      ENV: "PROD"
    ports:
      - "80:80"
      - "443:443"
    volumes:
      # for persistening SSL cert across deploys
      - type: bind
        target: /var/www/.cache
        source: /var/www/.cache
      # for persisting logs across deploys
      - type: bind
        target: /awillettebackend/logging
        source: /home/ubuntu
```

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

The terraform script also includes details for an ssh key. This is a public-key associated with a private-key on my local machine, the SSH connection comes into play later. Port ingress/egress rules are also declared for ssh, http, and https.

In the shell, executing `terraform plan && terraform apply` with the above script defined in the current directory as `website.tf` (\*.tf is valid) will deploy the EC2 instance into AWS.

## NoIP DNS Registration
[noip.com](https://www.noip.com/) is used to register an [A Record](https://en.wikipedia.org/wiki/List_of_DNS_record_types) for `*.andrewwillette.com`. The record points to the public IPv4 address of my now-deployed EC2 instance. Anytime the EC2 instance is re-deployed, this does have to be updated. This is seldom done though because redeploys are at the docker-container level not the EC2 level.

## Docker over SSH

I package and deploy my website as a docker container. Below is the `Dockerfile`.
```
FROM golang:latest
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build -o andrewwillettedotcom .
EXPOSE 80
EXPOSE 443
CMD ["./andrewwillettedotcom", "serve"]
```

The final key step is to configure docker commands on my local machine to execute on the docker-daemon of the recently-deployed EC2 instance. A [docker context](https://docs.docker.com/engine/context/working-with-contexts/) on my personal machine creates a connection to my EC2 instance's docker-daemon via SSH using the command `docker context create --docker host=ssh://ubuntu@<aws_public_ip> personalwebsite`. This is where the ssh-key from the terraform comes in! If that is configured correctly this command should "just work". It really is a great piece of docker I wasn't aware of prior to this effort.

I now have a single bash script I execute locally which will build the website's docker container from my local machine's code and deploy it to AWS.

```sh
#!/bin/sh
docker context use personalwebsite
docker-compose -f docker-compose-prod.yml down
docker-compose -f docker-compose-prod.yml build
docker-compose -f docker-compose-prod.yml up -d
```

Huzzah, happy coding!

