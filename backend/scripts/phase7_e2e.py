#!/usr/bin/env python3
"""Phase 7 E2E: notification channels, templates, send, delivery logs, and
alert-triggered notification linkage."""
import json, time, urllib.request, urllib.error

BASE = "http://127.0.0.1:8099/api/v1"

def req(method, path, body=None):
    data = json.dumps(body).encode() if body is not None else None
    r = urllib.request.Request(BASE + path, data=data, method=method,
                               headers={"Content-Type": "application/json"})
    try:
        resp = urllib.request.urlopen(r, timeout=10)
        raw = resp.read().decode()
        return resp.status, json.loads(raw) if raw else {}
    except urllib.error.HTTPError as e:
        return e.code, json.loads(e.read().decode())

def data_of(resp):
    return resp.get("data", {}) if isinstance(resp, dict) else {}

ok = True
def check(name, cond):
    global ok
    print(("PASS" if cond else "FAIL"), "-", name)
    if not cond: ok = False

# 1. List channels (seeded console channel should exist)
st, chs = req("GET", "/notification/channels")
ch_data = data_of(chs)
check("list channels", st == 200 and "items" in ch_data)
console_id = None
for c in ch_data.get("items", []):
    if c["type"] == "console":
        console_id = c["id"]
check("seeded console channel present", console_id is not None)

# 2. Create a webhook channel
st, created = req("POST", "/notification/channels", {
    "name": "Test Webhook", "type": "webhook", "is_enabled": True,
    "config": json.dumps({"url": "https://example.com/hook", "token": "x"}),
})
check("create webhook channel", st == 200 and data_of(created).get("id"))
wh_id = data_of(created).get("id")

# 3. Create a template for alert.fired
st, tpl = req("POST", "/notification/templates", {
    "name": "E2E Alert", "event_type": "alert.fired", "is_enabled": True,
    "subject": "[E2E] {{.Title}}", "body": "Val={{.Data.value}}",
    "channels": json.dumps(["console"]),
})
check("create template", st == 200 and data_of(tpl).get("id"))
tpl_id = data_of(tpl).get("id")

# 4. Send a test notification (alert.fired) - should log to delivery_logs
st, sent = req("POST", "/notification/send", {
    "event_type": "alert.fired", "title": "E2E Test Alert", "severity": "critical",
    "message": "disk full", "data": {"value": "96.5"},
})
check("send test notification", st == 200 and data_of(sent).get("dispatched") is True)

# 5. Delivery logs should contain the sent entry
time.sleep(0.5)
st, logs = req("GET", "/notification/delivery-logs")
log_data = data_of(logs)
check("list delivery logs", st == 200 and "items" in log_data)
found = any(l.get("event_type") == "alert.fired" for l in log_data.get("items", []))
check("alert.fired log present", found)

# 6. Bind the webhook channel to a rule and verify alert fires + notifies.
#    Create a rule on cpu.usage > 90 with the new webhook channel.
st, rule = req("POST", "/monitor/alert-rules", {
    "name": "E2E cpu>90", "metric_name": "cpu.usage", "condition": "gt",
    "threshold": 90, "severity": "critical", "duration": 0, "cooldown": 60,
    "is_enabled": True, "notification_channels": [wh_id],
})
check("create alert rule with channel", st == 200 and data_of(rule).get("id"))
rule_id = data_of(rule).get("id")

# 7. Ingest a cpu.usage > 90 sample -> should create a firing alert AND notify
st, ing = req("POST", "/monitor/metrics/ingest", {"metrics": [
    {"server_id": "synth-e2e", "metric_name": "cpu.usage", "value": 97.5, "unit": "%"},
]})
check("ingest high cpu", st == 200)

time.sleep(0.5)
st, alerts = req("GET", "/monitor/alerts")
al_data = data_of(alerts)
firing = [a for a in al_data.get("items", []) if a.get("metric_name") == "cpu.usage" and a.get("value", 0) > 90]
check("alert fired from ingest", len(firing) > 0)

st, logs2 = req("GET", "/notification/delivery-logs")
log2_data = data_of(logs2)
notified = any(l.get("channel_type") == "webhook" for l in log2_data.get("items", []))
check("webhook channel attempted delivery on alert", notified)

# 8. Cleanup
req("DELETE", f"/monitor/alert-rules/{rule_id}")
req("DELETE", f"/notification/channels/{wh_id}")
req("DELETE", f"/notification/templates/{tpl_id}")

print("=== PHASE7 E2E", "PASS" if ok else "FAIL", "===")
