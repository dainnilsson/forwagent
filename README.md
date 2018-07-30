# forwagent - gpg-agent forwarder
Forwagent allows forwarding gpg-agent sockets on Windows for use with GPG and
SSH over TCP. For example, allow WSL or VMs to create signatures or SSH using
keys stored on a YubiKey, without having access to the YubiKey itself.
Tunneled connections use an encrypted transport, with mutual authentication of
client and server. The server works on Windows, only. The client has only been
tested on Ubuntu, but might work on other distros.

NOTE: This project should be considered experimental, and is provided as-is.

## Quick setup
Both client and server need to have GnuPG installed, and their versions must be
close to ensure compatbility of the gpg-agent protocol.

### Server setup (Windows)
* The users private key should be usable from Windows.
* The `gpg-connect-agent.exe` executable should be in the users `PATH`.
* The `gpg-agent.conf` should have `enable-putty-support`.

Run the `forwagent-server.exe` executable to start the server. On first run,
this will generate a server key pair, and display the public portion. This is
the servers public key, save it somewhere (it's displayed each time you run the
server). By default the server runs on `127.0.0.1:4711`, but this can be set by
providing an IP:PORT string as an argument to the executable.

You'll most likely want to automate the startup of the server so that it runs
each time the computer starts. See the files in `doc/` for suggestions on how
to do this.

### Client setup (Ubuntu)
Run the `forwagent` binary to start the client. On first run, this will
generate a client key pair, and display the public portion. This is the clients
public key, which needs to be added to the server to allow connections. By
default, the client will attempt to connect to a server on `127.0.0.1:4711`,
but this can be configured by providing an IP:PORT string as an argument to the
executable.

When the client is run, two unix domain socket files are created in
`~/.gnupg/`, named `S.gpg-agent` and `S.gpg-agent.ssh`. These can be used by
`gpg` and `ssh`, and will be tunneled to the sockets on the server. You'll need
to configure SSH to use the socket:

  $ export SSH_AUTH_SOCK="$HOME/.gnupg/S.gpg-agent.ssh"

You'll most likely want to automate the startup of the client so that it runs
each time the computer starts. See the files in `doc/` for suggestions on how
to do this.

### Authentication (client and server)
With both server and client set up, you should have public keys for each of
them. You should also have a directory named `.forwagent` in the users home
directory on both systems.

In the servers `.forwagent` directory, create a file named `clients.allowed`
and paste the client public key. To allow multiple clients, add multiple keys,
one per line.

In the clients `.forwagent` directory, create a file named `servers.allowed`
and paste the servers public key. To allow connecting to multiple servers
(though not more than one at a time), add multiple keys, one per line.

### Usage
Once set up, you should be able to run `gpg` commands on the client machine,
and private keys from the server should be used. Note that the client
installation won't have the same public keyring as the server, so you'll need
to add your public key to it.

You should also be able to list your GPG authentication key under SSH by
running:

  $ ssh-add -L

And the `ssh` command should pick it up and use it automatically.

NOTE: If gpg requires a PIN or passphrase for any action, this will be prompted
for on the server, not the client.


## Listening on a different interface
By default the loopback interface is used, which works fine for WSL, but may
not be suitable for use with VMs. To limit attack surface, I wouldn't recommend
listening on a publically facing interface. An alternative is to listen on a
virtual interface accessible to both the host and VM.
