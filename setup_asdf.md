## asdf

### Install
``` sh
git clone https://github.com/asdf-vm/asdf.git ~/.asdf
echo -e "\n. $HOME/.asdf/asdf.sh" >> ~/.bashrc
echo -e '\n. $HOME/.asdf/completions/asdf.bash' >> ~/.bashrc
source ~/.bashrc
```

### Version Confirm
``` sh
asdf version
```

### Plugin Install(From .tool-versions)
``` sh
asdf install
```

### Plugin List
``` sh
asdf plugin-list-all
```

### Plugin Add
``` sh
asdf plugin add <Plugin>
```

### Installed Plugin List
``` sh
asdf plugin list
```

### Plugin Remove
``` sh
asdf plugin remove <Plugin>
```

### Available Version List
``` sh
asdf list all <Plugin>
```

### Plugin Install With Signatured Version
``` sh
asdf install <Plugin> <Version>
```
