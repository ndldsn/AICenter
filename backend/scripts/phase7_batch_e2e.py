#!/usr/bin/env python3
"""Phase 7.2 Batch Operations E2E: run a command across selected servers."""
import json, uuid, urllib.request, urllib.error, time

BASE = "http://127.0.0.1:8100/api/v1"

def req(method, path, body=None):
    data = json.dumps(body).encode() if body is not None else None
    r = urllib.request.Request(BASE + path, data=data, method=method,
                               headers={"Content-Type": "application/json"})
    try:
        resp = urllib.request.urlopen(r, timeout=20)
        return resp.status, json.loads(resp.read().decode())
    except urllib.error.HTTPError as e:
        try:
            body = json.loads(e.read().decode())
        except Exception:
            body = e.read().decode()
        return e.code, body

ok = True
def check(name, cond):
    global ok
    print(("PASS" if cond else "FAIL"), "-", name)
    if not cond: ok = False

# 0. Register a localhost server so the batch has a real target to run against.
srv = {"name": "local-e2e", "host": "localhost", "port": 22,
        "username": "admin", "auth_type": "password", "password": ""}
st, created = req("POST", "/servers", srv)
sid = created.get("data", {}).get("id") or created.get("id")
check("register localhost target", st == 201 and sid is not None)
if not sid:
    print("=== PHASE7.2 BATCH E2E FAIL ===")
    raise SystemExit(0)

# Ensure it's enabled/online-ish (no real SSH needed for localhost exec path).
# 1. Run a command batch against that server.
st, res = req("POST", "/servers/batch/command", {
    "command": "echo BATCH_PHASE7.2_OK",
    "server_ids": [sid],
    "timeout_seconds": 10,
})
check("batch command HTTP 200", st == 200)
items = res.get("data", {}).get("items") or res.get("items") or []
check("batch returned one result", len(items) == 1)
ok_out = any("BATCH_PHASE7.2_OK" in (it.get("stdout") or "") for it in items)
check("command output echoed on host", ok_out)
fail_out = any(it.get("status") == "ok" for it in items)
check("result status ok", fail_out)

# 2. Run a command that exits non-zero to verify exit code propagation.
st2, res2 = req("POST", "/servers/batch/command", {
    "command": "exit 7", "server_ids": [sid], "timeout_seconds": 10,
})
items2 = res2.get("data", {}).get("items") or res2.get("items") or []
check("non-zero exit code captured",
      any((it.get("exit_code") == 7 and it.get("status") == "ok") for it in items2))

# 3. Run a command that times out (sleep exceeds timeout).
st3, res3 = req("POST", "/servers/batch/command", {
    "command": "sleep 10", "server_ids": [sid], "timeout_seconds": 1,
})
items3 = res3.get("data", {}).get("items") or res3.get("items") or []
check("timeout surfaces as failure",
      any(it.get("status") == "failed" and "timeout" in (it.get("error") or "") for it in items3))

print("=== PHASE7.2 BATCH E2E", "PASS" if ok else "FAIL", "===")
