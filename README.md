# go-speed-dial

A GO project replicating and building upon the CLI tool made in the repository [speed-dial](https://github.com/alexanderConstantinescu/speed-dial). Basically this tool is intented as an intro for me to the GO programming language. 

## Installation 

### Linux - based system

```
curl https://raw.githubusercontent.com/alexanderConstantinescu/go-speed-dial/master/install.sh >> tmp.sh && chmod +x tmp.sh && sudo ./tmp.sh && rm tmp.sh
```

The linux installation also does a setup of bash completion for speed dial during the install. The script needs to be executed as root.   

### Windows - with UNIX terminal emulator 

**Run as administrator**
```
curl https://raw.githubusercontent.com/alexanderConstantinescu/go-speed-dial/master/install.sh >> tmp.sh && chmod +x tmp.sh && ./tmp.sh && rm tmp.sh
```

**Attention**: syscall.Exec seems to have an issue on windows, this is currently under investigation

## Usage:

Usage has been improved a bit, please view the following

### Save

```
speed-dial save -key "your-key" -val "command"
```

will save your command to a key in a .dial_keys file in your $HOME. The tool also allows for variable arguments to be associated to the command you save (this is done using the characters: {}, indicating variable argument), ex:

```
speed-dial save -key demo -val "ssh {}@ip"
```

which if later called as:

```
speed-dial demo user
```

will be interpreted as:

```
ssh user@ip
```

thus giving you the possibility to associate any username at execution.

**Attention:** the keyword "keys" is reserved and should not be used when saving commands. 

### Update

```
speed-dial update -key "your-key" -val "your-new-command"
```

will update your command to a key in a dialKeys.txt file. Pay attention to the quotes!

### Delete

```
speed-dial delete -key "your-key"
```

will delete your saved key

### List

```
speed-dial list
```

will list all your saved commands

### Export 


```
speed-dial export -ip ${IP} -id ${IDENTITY_FILE} - user ${USER}
```

Export allows you to export the .dial_keys file to any remote server. This is useful in case you already have the binary installed on a remote machine and want to export your preferences. 

### Execute

```
speed-dial key
```

will execute your saved command

## Note

* You might want to associate an alias for the binary as to more easily launch it, do so change .bash_aliases or your .bashrc in your $HOME directory. 
* Only runs on a UNIX based terminal / terminal emulator. 
