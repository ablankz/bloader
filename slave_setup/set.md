``` sh
terraform -chdir=slave_setup/ec2 apply
```

```sh
ssh -i ssh_keys/slave.id_rsa ec2-user@{{ip address}}
```