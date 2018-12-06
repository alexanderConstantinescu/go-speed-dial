# go-speed-dial

A Go project replicating and building upon the CLI tool made in the repository [speed-dial](https://github.com/alexanderConstantinescu/speed-dial). Basically this tool is intented as an intro for me to the Go programming language. 

This tool is nothing but a binary replicating the built-in BASH command: "alias" with some added sugar (features). 

This binary could be useful for you in the following scenarios: 

* you feel that alias is a bit limited (having to manually add commands to your ~/.bash_aliases file for "persistant storage", not being able to export it easily to other machines/virtual environments/etc)
* you'd like to continue using aliases in shells which do not support it (lightweight containers running standard old-school shell, for example)
* you manage clusters of containers and would like to have access to the same aliases in all containers (for debugging purposes, for example) and like to have them synchronized.

## Installation 

### Linux - based system

```
curl https://raw.githubusercontent.com/alexanderConstantinescu/go-speed-dial/master/install.sh >> tmp.sh && chmod +x tmp.sh && sudo ./tmp.sh && rm tmp.sh
```

The linux installation also does a setup of bash completion for speed dial during the install. The script needs to be executed as root.   

### Windows  

This tool cannot be used on windows as the OS does not have an implementation of the execve system call in linux. This tool relies fundamentally on this system call as to replace the sd process with the user command requested during its execution. The best you can do is use the tool as a reminder of commands. To install:

**Run as administrator**
```
curl https://raw.githubusercontent.com/alexanderConstantinescu/go-speed-dial/master/install.sh >> tmp.sh && chmod +x tmp.sh && ./tmp.sh && rm tmp.sh
```

## Usage:

Usage has been improved a bit, please view the following

### Save

```
speed-dial save -key "your-key" -val "command"
```

will save your command to a key in a .dial_keys file in your $HOME. The tool also allows for variable arguments to be associated to the command you save (this is done using the characters: {}, indicating variable argument), ex:

```
speed-dial save -key demo -val "ssh {1}@ip"
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

You can also "complete" a command by saving a key as follows:

```
sd save -key print -val "echo "
```

and invoking it as

```
$ sd print hello world 
hello world
```

This works just as the standard built-in: "alias"

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

Alternatively you could use an defined ssh alias to export the file, as such you will be able to perform a multi-hop export as well.

```
speed-dial export -ssh $SSH_ALIAS
```

### Execute

```
speed-dial key
```

will execute your saved command

## Note

* You might want to associate an alias for the binary as to more easily launch it, do so change .bash_aliases or your .bashrc in your $HOME directory. 
* Only runs on a Linux based terminal / terminal emulator. 
