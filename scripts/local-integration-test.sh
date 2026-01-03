#!/bin/sh
set -e

PORT=80
SERVER_PID=""

cleanup() {
    if [ -n "$SERVER_PID" ]; then
        echo "local-integration-test: Stopping server (PID: $SERVER_PID)..."
        kill "$SERVER_PID" 2>/dev/null || true
        wait "$SERVER_PID" 2>/dev/null || true
    fi
}

trap cleanup EXIT

# prepends message with local-integration-test:
print_integration_test_message() {
    echo "local-integration-test: $1"
}

print_integration_test_message "Building application..."
go build -o andrewwillettedotcom .

print_integration_test_message "Starting server on port $PORT..."
./andrewwillettedotcom serve &
SERVER_PID=$!

# Wait for server to be ready
print_integration_test_message "Waiting for server to start..."
for i in $(seq 1 30); do
    if curl -s -o /dev/null -w "" "http://localhost:$PORT/" 2>/dev/null; then
        print_integration_test_message "Server is ready"
        break
    fi
    if [ "$i" -eq 30 ]; then
        print_integration_test_message "ERROR: Server failed to start within 30 seconds"
        exit 1
    fi
    sleep 1
done

FAILED=0

test_endpoint() {
    local path="$1"
    local expected_status="$2"
    local description="$3"

    status=$(curl -s -o /dev/null -w "%{http_code}" "http://localhost:$PORT$path")
    if [ "$status" -eq "$expected_status" ]; then
        print_integration_test_message "✓ $description ($path) - $status"
    else
        print_integration_test_message "✗ $description ($path) - expected $expected_status, got $status"
        FAILED=1
    fi
}

echo ""
print_integration_test_message "Testing endpoints..."
print_integration_test_message "===================="

test_endpoint "/" 200 "Homepage"
test_endpoint "/music" 200 "Music page"
test_endpoint "/sheet-music" 200 "Sheet music page"
test_endpoint "/blog" 200 "Blog listing"
test_endpoint "/key-of-the-day" 200 "Key of the day"
test_endpoint "/static/main.css" 200 "CSS file"
test_endpoint "/robots.txt" 200 "Robots.txt"

echo ""
if [ "$FAILED" -eq 0 ]; then
    print_integration_test_message "All tests passed!"
    exit 0
else
    print_integration_test_message "Some tests failed!"
    exit 1
fi
