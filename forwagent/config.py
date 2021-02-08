from .common import CONF_DIR, KEY, CERT, TRUSTED
from cryptography import x509
from cryptography.x509.oid import NameOID
from cryptography.hazmat.primitives import serialization, hashes
from cryptography.hazmat.primitives.asymmetric import rsa
import os
import datetime
import socket


def generate_pair():
    key = rsa.generate_private_key(65537, 4096)
    name = socket.gethostname() + "-" + os.urandom(4).hex()
    issuer = x509.Name([x509.NameAttribute(NameOID.COMMON_NAME, name)])
    cert = (
        x509.CertificateBuilder()
        .subject_name(issuer)
        .issuer_name(issuer)
        .public_key(key.public_key())
        .serial_number(x509.random_serial_number())
        .not_valid_before(datetime.datetime.utcnow())
        .not_valid_after(datetime.datetime.utcnow() + datetime.timedelta(days=3650))
        .add_extension(x509.BasicConstraints(ca=True, path_length=None), critical=True)
        .sign(key, hashes.SHA256())
    )
    return key, cert


def init():
    if not os.path.isdir(CONF_DIR):
        print("Creating directory:", CONF_DIR)
        os.mkdir(CONF_DIR, mode=0o700)

    if not os.path.isfile(KEY) and not os.path.isfile(CERT):
        print("Generating key and certificate...")
        key, cert = generate_pair()
        with open(KEY, "wb") as f:
            f.write(
                key.private_bytes(
                    serialization.Encoding.PEM,
                    serialization.PrivateFormat.PKCS8,
                    serialization.NoEncryption(),
                )
            )
        print("Wrote key to:", KEY)

        with open(CERT, "wb") as f:
            f.write(cert.public_bytes(serialization.Encoding.PEM))
        print("Wrote cert to:", CERT)

    if not os.path.isfile(TRUSTED):
        print("Creating empty CA bundle:", TRUSTED)
        with open(TRUSTED, "wb") as f:
            f.write(b"")
        print("You will need to add at least 1 certificate to this file.")

    print("Setup complete.")
