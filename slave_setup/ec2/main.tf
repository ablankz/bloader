resource "aws_instance" "ec2_instance" {
  count         = var.instance_count
  ami           = data.aws_ssm_parameter.amazonlinux_2.value
  instance_type = var.instance_type
  subnet_id     = aws_subnet.public.id
  key_name      = aws_key_pair.ssh_key_pair.key_name
  vpc_security_group_ids = [
    aws_security_group.grpc_sg.id
  ]

  user_data = <<-EOF
#!/bin/bash
cd /home/ec2-user
sudo yum update -y
sudo yum install -y git
git --version

ARCH=$(uname -m)
if [ "$ARCH" = "x86_64" ]; then
  curl -LO https://go.dev/dl/go1.23.4.linux-amd64.tar.gz
  sudo tar -C /usr/local -xzf go1.23.4.linux-amd64.tar.gz
  echo 'export PATH=$PATH:/usr/local/go/bin' >> /home/ec2-user/.bashrc
  source /home/ec2-user/.bashrc
  go version
elif [ "$ARCH" = "aarch64" ]; then
  curl -LO https://go.dev/dl/go1.23.4.linux-arm64.tar.gz
  sudo tar -C /usr/local -xzf go1.23.4.linux-arm64.tar.gz
  echo 'export PATH=$PATH:/usr/local/go/bin' >> /home/ec2-user/.bashrc
  source /home/ec2-user/.bashrc
  go version
else
  echo "Unsupported architecture: $ARCH"
fi

git clone https://github.com/ablankz/bloader.git
cd bloader

if [ "${var.tls_enabled}" = true ]; then
  # Conditionally add TLS private key
  mkdir -p $(dirname "${var.slave_ca_cert_file_path}")
  mkdir -p $(dirname "${var.slave_ca_key_file_path}")
  mkdir -p $(dirname "${var.slave_cert_file_path}")
  mkdir -p $(dirname "${var.slave_key_file_path}")
  cat <<EOF_TLS_CERT > ${var.slave_ca_cert_file_path}
${tls_self_signed_cert.ca_cert[0].cert_pem}
EOF_TLS_CERT
  cat <<EOF_TLS_KEY > ${var.slave_ca_key_file_path}
${tls_private_key.ca_key[0].private_key_pem}
EOF_TLS_KEY
  # Generate Slave private key
  openssl genrsa -out ${var.slave_key_file_path} 2048

  # Create a CSR for the slave certificate
  openssl req -new -key ${var.slave_key_file_path} -subj "/CN=localhost" -out /tmp/slave.csr

  # Generate the Slave certificate using the CA key and certificate
  openssl x509 -req \
    -in /tmp/slave.csr \
    -CA ${var.slave_ca_cert_file_path} \
    -CAkey ${var.slave_ca_key_file_path} \
    -CAcreateserial \
    -out ${var.slave_cert_file_path} \
    -days 365 \
    -sha256

  # Cleanup
  rm -f /tmp/slave.csr

  export HOME=/home/ec2-user

  go env

  go run main.go slave run --config bloader/slave_config.yaml
fi
EOF

  tags = {
    Name = "bloader-slave-instance-${count.index + 1}"
  }
}

resource "aws_eip" "instance_eip" {
  count       = var.instance_count
  instance    = aws_instance.ec2_instance[count.index].id
  domain   = "vpc"

  tags = {
    Name = "bloader-eip-${count.index + 1}"
  }
}