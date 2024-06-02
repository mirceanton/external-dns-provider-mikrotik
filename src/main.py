from flask import Flask, request, jsonify
from utils.mikrotik import MikrotikAPI

mikrotik = MikrotikAPI()
mikrotik.connect()

app = Flask(__name__)

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
