# forwagent - gpg-agent forwarder
Forwagent allows forwarding gpg-agent sockets on Windows for use with GPG and
SSH over TCP. For example, allow WSL or VMs to create signatures or SSH using
keys stored on a YubiKey, without having access to the YubiKey itself.
Tunneled connections use an encrypted transport, with mutual authentication of
client and server. The server works on Windows, only. The client has only been
tested on Ubuntu, but might work on other distros.

NOTE: This project should be considered experimental, and is provided as-is.

NOTE: This project was initially written in Go. The old code can be found in
the git history.


## Quick setup
Both client and server need to have GnuPG installed, and their versions should
be close to ensure compatbility of the gpg-agent protocol.

You can use `pip` (or an alternative, like `pipx`) to install `forwagent`. The
package defines two "extras" which are needed to run the server, and to
initialize the configuration.

To install for running as a server (or client):

    $ pip install forwagent[server]

To install for running as a client, with the ability to initialize
configuration:

    $ pip install forwagent[init]

To install for running as a client, with no dependencies (with existing
configuration):

    $ pip install forwagent


### Server setup (Windows)
* The users private key should be usable from Windows.
* The `gpg-connect-agent.exe` executable should be in the users `PATH`.
* The `gpg-agent.conf` should have `enable-putty-support`.

Initialize the configuration by running `forwagent init`. This will create a
`.forwagent/` directory in your users $HOME, and a self-signed certificate
which will be used in the files `key.pem` and `cert.pem`. An empty
`trusted.pem` file will be created for adding trusted client certificates.

You will need to add client certificates to the `trusted.pem` file before the
server will accept connections. If any of the files in `.forwagent/` are changed,
the server will need to be restarted for those changes to take effect.

Run `forwagent server` to start the server. By default the server runs on
`127.0.0.1:4711`, but this can be set by providing the --interface and --port
arguments when running.

You'll most likely want to automate the startup of the server so that it runs
each time the computer starts. See the files in `doc/` for suggestions on how
to do this.


### Client setup (Ubuntu)
Initialize the configuration by running `forwagent init`. This will create a
`.forwagent/` directory in your users $HOME, and a self-signed certificate
which will be used in the files `key.pem` and `cert.pem`. An empty
`trusted.pem` file will be created for adding the servers certificate.

Run `forwagent agent` to start the client. By default the client will connect
to a server running on `127.0.0.1:4711`, but this can be set by providing the
--interface and --port arguments when running.

When the client is run, two unix domain socket files are created in
`~/.gnupg/`, named `S.gpg-agent` and `S.gpg-agent.ssh` (the names and locations
of these may be different, refer to the output of `gpgconf --list-dirs` for the
paths used). These can be used by `gpg` and `ssh`, and will be tunneled to the
sockets on the server. You'll need to configure SSH to use the socket:

    $ export SSH_AUTH_SOCK=$(gpgconf --list-dirs agent-ssh-socket)

You'll most likely want to automate the startup of the client so that it runs
each time the computer starts. See the files in `doc/` for suggestions on how
to do this.


### Authentication (client and server)
Mutual authentication is done by using client and server X.509 certificates.
This is done by copying the `.forwagent/cert.pem` file from the server, to the
`.forwagent/trusted.pem` of the client, and vice versa.
To allow multiple clients (or servers) the `trusted.pem` file can contain
multiple certificates.


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
