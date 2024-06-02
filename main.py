from flask import Flask, request, jsonify

app = Flask(__name__)

@app.route('/dns', methods=['GET','POST'])
def update_dns():
    data = request.json
    print(data)
    # TODO
    return jsonify({"status": "success"})

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=8088)
