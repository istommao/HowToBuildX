import os
import time

import io
import qrcode

from cryptography.hazmat.primitives.twofactor.totp import TOTP
from cryptography.hazmat.primitives.hashes import SHA1

def generate_totp(key=None, length=6, algorithm=SHA1(), time_step=30):
    if not key:
        key = os.urandom(20)

    totp = TOTP(key, length, algorithm, time_step)
    return totp


def print_totp_qrcode(data):
    qr = qrcode.QRCode()
    qr.add_data(data)
    f = io.StringIO()
    qr.print_ascii(out=f)
    f.seek(0)
    print(f.read())

    f.close()


def main():
    totp = generate_totp()
    time_value = time.time()
    totp_value = totp.generate(time_value)

    totp.verify(totp_value, time_value)

    account_name = 'totp@codemax.com'
    issuer_name = 'CodeMax Inc'
    data = totp.get_provisioning_uri(account_name, issuer_name)
    print("TOTP:\n", data)
    print("QRCODE use Google Authenticator:")
    print_totp_qrcode(data)


if __name__ == '__main__':
    main()
