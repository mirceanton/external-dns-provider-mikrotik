import routeros_api
import sys
import os


class MikrotikAPI:
    STATIC_DNS_RESOURCE_PATH="/ip/dns/static"
    api: any
    connection: any
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

    def __del__(self):
        if self.connection is not None:
            self.connection.disconnect()

    def connect(self):
        try:
            self.connection = routeros_api.RouterOsApiPool(
                self.host,
                username=self.username,
                password=self.password,
                port=self.port,
                use_ssl=self.use_ssl,
                ssl_verify=self.ssl_verify,
                ssl_verify_hostname=self.ssl_verify,
                plaintext_login=True,
            )
            self.api = self.connection.get_api()
            print("Connection successful!")
        except Exception as e:
            print(f"Failed to connect to the router: {e}")
            sys.exit(1)

    def add_dns_record(self, fqdn: str, ip: str) -> bool:
        try:
            self.api.get_resource(self.STATIC_DNS_RESOURCE_PATH).add(
                name=fqdn, address=ip
            )
            print(f"Added DNS record: {fqdn} -> {ip}")
            return True
        except Exception as e:
            print(f"Error adding DNS record: {e}")
            return False

    def update_dns_record(self, fqdn: str, ip: str) -> bool:
        try:
            dns_resource = self.api.get_resource(self.STATIC_DNS_RESOURCE_PATH)
            existing_record = dns_resource.get(name=fqdn)

            if existing_record:
                dns_resource.set(id=existing_record[0]['id'], address=ip)
                print(f"Updated DNS record: {fqdn} -> {ip}")
            else:
                self.add_dns_record(fqdn, ip)
            return True
        except Exception as e:
            print(f"Error updating DNS record: {e}")
            return False

    def delete_dns_record(self, fqdn: str) -> bool:
        try:
            dns_resource = self.api.get_resource(self.STATIC_DNS_RESOURCE_PATH)
            existing_record = dns_resource.get(name=fqdn)
            if existing_record:
                dns_resource.remove(id=existing_record[0]['id'])
                print(f"Deleted DNS record: {fqdn}")
            return True
        except Exception as e:
            print(f"Error deleting DNS record: {e}")
            return False
