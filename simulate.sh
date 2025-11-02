#!/bin/bash
RED='\033[0;31m'
CYAN='\033[0;36m'
BOLD='\033[1m'
UNDERLINE='\033[4m'
RESET='\033[0m'

set -euo pipefail

cleanup() {
    echo "Cleaning up..."
    [ -n "${TAIL_PID:-}" ] && kill "$TAIL_PID" 2>/dev/null || true
    [ -n "${SERVER_PID:-}" ] && kill "$SERVER_PID" 2>/dev/null || true
    deactivate 2>/dev/null || true
    rm -rf venv
}
trap cleanup EXIT

# start venv
python3 -m venv venv
source venv/bin/activate

# install server deps
pip install -r ./server/requirements.txt

# start python server and log output
mkdir -p logs
python3 ./server/server.py > logs/server.log 2>&1 &
SERVER_PID=$!


# give server a moment to start
sleep 1

# start clients (will run while server output is printed)
# Simple client
echo -e "$CYAN$BOLD$UNDERLINE Simple Client Output: $RESET"
go run cmd/simple/main.go

# Wait Group client
echo -e "\n$CYAN$BOLD$UNDERLINE Wait Group Client Output: $RESET"
go run cmd/waitgroups/main.go    

# Fan-out client
echo -e "\n$CYAN$BOLD$UNDERLINE Fan-out/Fan-in Client Output: $RESET"
go run cmd/fanoutin/main.go

# Fan-out Fan-in with Backpressure client
echo -e "\n$CYAN$BOLD$UNDERLINE Fan-out/Fan-in with Backpressure Client Output: $RESET"
go run cmd/fanoutinwbp/main.go