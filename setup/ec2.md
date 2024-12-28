# AWS EC2 Setup

``` sh
sudo dnf update -y
sudo dnf install -y git
git --version
```

amd64
``` sh
curl -LO https://go.dev/dl/go1.23.4.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.23.4.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
go version
```

arm64
``` sh
curl -LO https://go.dev/dl/go1.23.4.linux-arm64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.23.4.linux-arm64.tar.gz
export PATH=$PATH:/usr/local/go/bin
go version
```

``` sh
git clone https://github.com/ablankz/bloader.git
```