#!/usr/bin/env python3
"""Phase 7.1 Web Terminal E2E: PTY session over WebSocket runs a real command."""
import json, time, urllib.request, urllib.error, threading, base64

try:
    import websocket  # websocket-client
except ImportError:
    websocket = None

BASE = "http://127.0.0.1:8100/api/v1"
WS = "ws://127.0.0.1:8100/ws/terminal"

def req(method, path, body=None):
    data = json.dumps(body).encode() if body is not None else None
    r = urllib.request.Request(BASE + path, data=data, method=method,
                               headers={"Content-Type": "application/json"})
    try:
        resp = urllib.request.urlopen(r, timeout=10)
        return resp.status, json.loads(resp.read().decode())
    except urllib.error.HTTPError as e:
        return e.code, json.loads(e.read().decode())

ok = True
def check(name, cond):
    global ok
    print(("PASS" if cond else "FAIL"), "-", name)
    if not cond: ok = False

if websocket is None:
    print("SKIP - websocket-client not installed")
    raise SystemExit(0)

# 1. Create a terminal session
st, created = req("POST", "/terminal/sessions", {"cols": 80, "rows": 24})
sid = created.get("session_id") or created.get("data", {}).get("session_id")
check("create terminal session", st == 200 and sid is not None)

# 2. Connect over WS and run a command
received = []
def on_msg(ws, raw):
    try:
        received.append(raw)
    except Exception:
        pass

ws = websocket.WebSocket()
ws.connect(WS + "?session=" + sid, timeout=10)
ws.send(json.dumps({"type": "input", "data": "echo HELLO_TERMINAL_E2E\n"}))
deadline = time.time() + 6
while time.time() < deadline:
    try:
        ws.settimeout(1.0)
        chunk = ws.recv()
        if isinstance(chunk, bytes):
            chunk = chunk.decode(errors="replace")
        received.append(chunk)
    except Exception:
        if any("HELLO_TERMINAL_E2E" in r for r in received):
            break
    if any("HELLO_TERMINAL_E2E" in r for r in received):
        break

output = "".join(received)
check("pty echoed command output", "HELLO_TERMINAL_E2E" in output)

# 3. List sessions shows our session (while the WS is still connected).
st, lst = req("GET", "/terminal/sessions")
items = lst.get("data", {}).get("items") or lst.get("items") or []
check("list terminal sessions", st == 200 and any(s.get("id") == sid for s in items))

# 4. Close session
ws.close()
st, _ = req("POST", f"/terminal/sessions/{sid}/close")
check("close terminal session", st == 200)

print("=== PHASE7.1 TERMINAL E2E", "PASS" if ok else "FAIL", "===")
