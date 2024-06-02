from flask import Flask, request, jsonify
import routeros_api
import os
import sys


# =================================================================================================
# Set up Mikrotik connection
# =================================================================================================
MIKROTIK_HOST = os.getenv('MIKROTIK_HOST')
MIKROTIK_PORT = os.getenv('MIKROTIK_PORT', '8728')
if not MIKROTIK_HOST:
    print("Environment variable MIKROTIK_HOST is not set")
    sys.exit(1)

MIKROTIK_USER = os.getenv('MIKROTIK_USER')
if not MIKROTIK_USER:
    print("Environment variable MIKROTIK_USER is not set")
    sys.exit(1)

MIKROTIK_PASS = os.getenv('MIKROTIK_PASS')
if not MIKROTIK_PASS:
    print("Environment variable MIKROTIK_PASS is not set")
    sys.exit(1)

MIKROTIK_USE_SSL = os.getenv('MIKROTIK_USE_SSL', 'false').lower() in ('true', '1', 'yes')
MIKROTIK_SSL_VERIFY = os.getenv('MIKROTIK_SSL_VERIFY', 'false').lower() in ('true', '1', 'yes')

# Connect to the router
try:
    connection = routeros_api.RouterOsApiPool(
        MIKROTIK_HOST,
        username=MIKROTIK_USER,
        password=MIKROTIK_PASS,
        port=MIKROTIK_PORT,
        use_ssl=MIKROTIK_USE_SSL,
        ssl_verify=MIKROTIK_SSL_VERIFY,
        ssl_verify_hostname=MIKROTIK_SSL_VERIFY,
        plaintext_login=True,
    )
    routeros_api = connection.get_api()
except Exception as e:
    print(f"Failed to connect to the router: {e}")
    sys.exit(1)



# =================================================================================================
# Set up Flask API
# =================================================================================================
app = Flask(__name__)

@app.route('/dns', methods=['GET','POST'])
def update_dns():
    data = request.json
    print(data)
    # TODO
    return jsonify({"status": "success"})

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=8088)
