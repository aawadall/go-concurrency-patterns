"""A simple server implementation, simulating latency and errors."""
import random
import time
from flask import Flask, jsonify, request

# Latency and Error State
class ServerState:
    def __init__(self):
        self.latency = 0
        self.error_rate = 0.01  # 1% chance of error
        self.error_state = False
        self.last_request_time = time.time()

    def reset(self):
        self.latency = 0
        self.error_state = False

    def simulate_latency(self):
        
        print(f"Simulating latency: {self.latency:.3f} seconds")
        time.sleep(self.latency)

    def simulate_error(self):
        return self.error_state

    def update_state(self):
        """
        Update latency and error state using random walk.
        guided by time since last request.
        """
        current_time = time.time()
        time_diff = current_time - self.last_request_time
        self.last_request_time = current_time
        
        # use time_diff to scale up or down in latency changes
        latency_scale = min(max(time_diff / 5.0, 0.1), 2.0)
        latency_change = random.uniform(-0.1, 0.1) * latency_scale
        self.latency = max(0.0, self.latency + latency_change)

        # use time_diff to scale error state changes, if recent increase chance of error
        error_scale = min(max(time_diff / 10.0, 0.1), 2.0)
        if random.random() < 0.2 * error_scale:
            self.error_state = not self.error_state

# Define endpoint handler 
def handle_request():
    state = ServerState()
    state.update_state()
    state.simulate_latency()

    # Simulate error with a 20% chance
    if state.simulate_error():
        return jsonify({"error": "Simulated server error"}), 500

    # Return successful response
    data = {"message": "Hello, World!"}
    return jsonify(data), 200

# Main server function
def main(host, port):
    app = Flask(__name__)

    @app.route('/data', methods=['GET'])
    def get_data():
        return handle_request()

    app.run(host=host, port=port)

if __name__ == "__main__":
    main("0.0.0.0", 5000)