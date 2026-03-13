import requests
import json
import time
import sys

BASE_URL = "http://localhost"
MARKET_API = f"{BASE_URL}/api/market"
GATEWAY_URL = f"{BASE_URL}/mcp-gateway"

def verify():
    print("--- 1. Login ---")
    login_resp = requests.post(f"{BASE_URL}/api/authz/login", json={
        "username": "admin",
        "password": "admin123"
    })
    if login_resp.status_code != 200:
        print(f"Login failed: {login_resp.text}")
        return
    
    token = login_resp.json().get("data", {}).get("token")
    headers = {"Authorization": f"Bearer {token}"}
    print("Login successful")

    print("\n--- 2. Create Instance (Direct Mode) ---")
    # Point to the echo server we just started on 9999
    # In docker-compose, localhost is host.docker.internal or we can use the host IP.
    # But here we are running everything locally (or in the environment).
    # Since the agent starts echo_server.go locally on port 9999, 
    # the backend (running in docker) should access it via host.docker.internal:9999
    
    instance_req = {
        "name": "Header Test Instance",
        "accessType": "direct",
        "mcpProtocol": "sse",
        "sourceType": "direct",
        "sourceConfig": json.dumps({"url": "http://host.docker.internal:9999"}),
        "headers": {
            "X-Test-Instance-Header": "Verified-By-Antigravity",
            "Custom-Auth": "SecretToken123"
        },
        "enabledToken": True,
        "tokens": [
            {"name": "test-token", "token": "verify-token-123"}
        ]
    }
    
    create_resp = requests.post(f"{MARKET_API}/instance/create", json=instance_req, headers=headers)
    if create_resp.status_code != 200:
        print(f"Create instance failed: {create_resp.text}")
        return
    
    instance_id = create_resp.json().get("data", {}).get("instanceId")
    print(f"Instance created: {instance_id}")

    print("\n--- 3. Verify via Gateway ---")
    # The gateway endpoint should be /mcp-gateway/v1/instances/{instanceId}/...
    # or depends on the gateway routing.
    # Usually it's /mcp-gateway/api/v1/sse/instances/{instanceId} or similar.
    # Let's try the standard path: /mcp-gateway/instances/{instanceId}/sse
    
    gateway_headers = {"Authorization": "Bearer verify-token-123"}
    # Standard MCP path for listing tools or something
    test_path = f"{GATEWAY_URL}/instances/{instance_id}/tools/list"
    
    print(f"Calling gateway: {test_path}")
    gw_resp = requests.get(test_path, headers=gateway_headers)
    
    if gw_resp.status_code != 200:
        print(f"Gateway call failed ({gw_resp.status_code}): {gw_resp.text}")
        # Note: If it's SSE it might behave differently, but echo_server returns JSON.
        # If it returns 200, check the JSON body which contains headers seen by echo_server.
    
    try:
        echo_data = gw_resp.json()
        seen_headers = echo_data.get("headers", {})
        
        expected = {
            "X-Test-Instance-Header": "Verified-By-Antigravity",
            "Custom-Auth": "SecretToken123"
        }
        
        success = True
        print("\nVerification Results:")
        for k, v in expected.items():
            actual = seen_headers.get(k)
            if actual == v:
                print(f"  [PASS] {k}: {v}")
            else:
                print(f"  [FAIL] {k}: expected {v}, got {actual}")
                success = False
        
        if success:
            print("\nSUCCESS: End-to-end verification passed!")
        else:
            print("\nFAILURE: Header mismatch.")
            
    except Exception as e:
        print(f"Error parsing gateway response: {e}")
        print(f"Response text: {gw_resp.text}")

if __name__ == "__main__":
    verify()
