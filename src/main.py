from flask import Flask, request, jsonify
from utils.mikrotik import MikrotikAPI

mikrotik = MikrotikAPI()
mikrotik.connect()

app = Flask(__name__)


# =================================================================================================
# KUBERNETES ENDPOINTS
# =================================================================================================
@app.route('/readiness', methods=['GET'])
def readiness():
    return jsonify({"status": "ready"}), 200

@app.route('/liveness', methods=['GET'])
def liveness():
    return jsonify({"status": "live"}), 200


# =================================================================================================
# EXTERNAL-DNS ENDPOINTS
# =================================================================================================
@app.route('/dns', methods=['GET','POST'])
def update_dns():
    data = request.json
    print(data)
    # TODO
    return jsonify({"status": "success"})


if __name__ == '__main__':
    app.run(host='0.0.0.0', port=8088)
