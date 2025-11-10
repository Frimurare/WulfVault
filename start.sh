#!/bin/bash
# Sharecare Startup Script
# Automatically detects server IP and starts with correct configuration

# Get the primary network IP (not localhost)
SERVER_IP=$(hostname -I | awk '{print $1}')

# Default to localhost if no IP found
if [ -z "$SERVER_IP" ]; then
    SERVER_IP="localhost"
fi

export SERVER_URL="http://$SERVER_IP:8080"

echo "Starting Sharecare with SERVER_URL=$SERVER_URL"

# Start the server
nohup ./sharecare > server.log 2>&1 &

echo "Server started! PID: $!"
echo "Waiting for startup..."
sleep 3

# Show startup logs
tail -20 server.log
