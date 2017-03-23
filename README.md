```
$> ecsy help
Wraps the ECS SDK in a more user-friendly (for me at least) way

Usage:
  ecsy [command]

Available Commands:
  add         Associates a .pem SSH key with a cluster, allowing SSH into EC2 instances
  help        Help about any command
  ssh         Secure Shell into one of the service container instances' EC2 host machines

Flags:
      --cluster string   specifies an ECS Cluster (if empty, will load menu)
      --config string    config file (default is $HOME/.ecsy.yaml)

Use "ecsy [command] --help" for more information about a command.
```

#### Installation (OSX only for now)

Example (v0.0.3)

```
wget -O /usr/local/bin/ecsy https://github.com/oberd/ecsy/releases/download/v0.0.3/ecsy-v0.0.3-darwin-amd64
chmod +x /usr/local/bin/ecsy
```


#### Usage

##### Adding SSH Keys

To get started, you will need to register some ssh keys (per cluster).

For example, if you have an ECS cluster called `my-app-dev`, whose instances
use the ssh key `my-app.pem`, register an ssh key with:

```
ecsy add my-app-dev ~/.ssh/my-app.pem
```

You only have to do this once, it will persist to `~/.ecsy.yaml` (by default)

##### Running commands

Most other help is available on the CLI.  Check it out, and good luck!
