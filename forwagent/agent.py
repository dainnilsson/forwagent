from .common import TYPE_SSH, TYPE_GPG, forward, run, KEY, CERT, TRUSTED
import socket
import ssl
import os
import logging

logger = logging.getLogger(__name__)


DOMAIN_SOCKETS = {
    "S.gpg-agent": TYPE_GPG,
    "S.gpg-agent.ssh": TYPE_SSH,
}


def main(server_addr):
    sockets = {}

    context = ssl.SSLContext(ssl.PROTOCOL_TLS_CLIENT)
    context.load_verify_locations(TRUSTED)
    context.load_cert_chain(CERT, KEY)
    context.check_hostname = False

    def forward_to(b):
        return lambda a: forward(sockets, a, b)

    def accept(preamble):
        def inner(sock):
            client, _ = sock.accept()
            logger.info("Client connected for %r", preamble)

            server = context.wrap_socket(
                socket.socket(socket.AF_INET, socket.SOCK_STREAM)
            )
            server.connect(server_addr)
            server.sendall(preamble)

            sockets[client] = forward_to(server)
            sockets[server] = forward_to(client)

        return inner

    socket_files = []
    for name, preamble in DOMAIN_SOCKETS.items():
        s = socket.socket(
            socket.AF_UNIX,
            socket.SOCK_STREAM,
        )
        fname = os.path.join(os.path.expanduser("~/.gnupg"), name)
        if os.path.exists(fname):
            logger.debug("Remove stale socket file: %s", fname)
            os.remove(fname)
        s.bind(fname),
        socket_files.append(fname)
        s.listen()
        logger.debug("Created domain socket: %s", fname)
        sockets[s] = accept(preamble)

    logger.info("Accepting incomming connection, targetting : %s:%d", *server_addr)

    try:
        run(sockets)
    finally:  # Clean up stale sockets
        for fname in socket_files:
            os.remove(fname)
