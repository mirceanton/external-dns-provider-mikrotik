import routeros_api
import sys
import os


class MikrotikAPI:
    api: any
    host: str
    port: int
    username: str
    password: str
    use_ssl: bool
    ssl_verify: bool


    def __init__(self):
        self.host = os.getenv('MIKROTIK_HOST')
        self.port = os.getenv('MIKROTIK_PORT', '8728')
        if not self.host:
            print("Environment variable MIKROTIK_HOST is not set")
            sys.exit(1)

        self.username = os.getenv('MIKROTIK_USER')
        if not self.username:
            print("Environment variable MIKROTIK_USER is not set")
            sys.exit(1)

        self.password = os.getenv('MIKROTIK_PASS')
        if not self.password:
            print("Environment variable MIKROTIK_PASS is not set")
            sys.exit(1)

        self.use_ssl = os.getenv('MIKROTIK_USE_SSL', 'false').lower() in ('true', '1', 'yes')
        self.ssl_verify = os.getenv('MIKROTIK_SSL_VERIFY', 'false').lower() in ('true', '1', 'yes')

    def connect(self):
        try:
            connection = routeros_api.RouterOsApiPool(
                self.host,
                username=self.username,
                password=self.password,
                port=self.port,
                use_ssl=self.use_ssl,
                ssl_verify=self.ssl_verify,
                ssl_verify_hostname=self.ssl_verify,
                plaintext_login=True,
            )
            self.api = connection.get_api()
            print("Connection successful!")
        except Exception as e:
            print(f"Failed to connect to the router: {e}")
            sys.exit(1)

    def add_dns_record(self, fqdn: str, ip: str) -> bool:
        pass

    def update_dns_record(self, fqdn: str, ip: str) -> bool:
        pass

    def delete_dns_record(self, fqdn: str, ip: str) -> bool:
        pass
