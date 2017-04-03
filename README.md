```
$> ecsy help
Wraps the ECS SDK in a more user-friendly (for me at least) way

Usage:
  ecsy [command]

Available Commands:
  add         Associates a .pem SSH key with a cluster, allowing SSH into EC2 instances
  env         Used to manage environment variables of service task definitions
  help        Help about any command
  self-update Update the ecsy cli binary on your system
  ssh         Secure Shell into one of the service container instances' EC2 host machines

Flags:
      --config string   config file (default is $HOME/.ecsy.yaml)

Use "ecsy [command] --help" for more information about a command.
```

#### Installation (OSX only for now)

Example (v0.0.8)

```
wget -O /usr/local/bin/ecsy https://github.com/oberd/ecsy/releases/download/v0.0.7/ecsy-v0.0.8-darwin-amd64
chmod +x /usr/local/bin/ecsy
```

#### Updating

You can run `ecsy self-update` to get the latest version

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
