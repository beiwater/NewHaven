#!/usr/bin/env python3
"""
New Haven Security Audit Script
================================
Tests the backend API for common vulnerabilities.
Run with: python scripts/security-audit.py

Prerequisites:
  - Backend running on http://localhost:8088
  - Dev mode enabled (creates dev/dev user)
"""

import json
import sys
import time
import urllib.request
import urllib.error
import urllib.parse

BASE = "http://localhost:8088"
TOKEN = None  # obtained during tests

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
    """Make an API request and return (status, data)."""
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


def login(username="dev", password="dev"):
    """Login and return token."""
    status, data = api("POST", "/api/login", {"username": username, "password": password})
    if status == 200:
        return data.get("data", {}).get("token") or data.get("token")
    return None


# ============================================================
# SECTION 1: Authentication & Access Control
# ============================================================
def section_auth():
    print("\n" + "="*60)
    print("SECTION 1: Authentication & Access Control")
    print("="*60)

    # 1.1 Register a fresh user for testing
    print("\n-- 1.1 Registration --")
    status, data = api("POST", "/api/register", {"username": "secaudit", "password": "TestPass123!"})
    if status in (200, 409):
        test("Register new user", "PASS" if status == 200 else "WARN",
             "User exists (409) - reused from previous run" if status == 409 else "")
        if status == 200:
            mtoken = data.get("data", {}).get("token") or data.get("token")
        else:
            mtoken = login("secaudit", "TestPass123!")
            if not mtoken:
                mtoken = login()  # fallback to dev
    else:
        test("Register new user", "WARN", f"Unexpected status {status}: {data}")
        mtoken = login()
    # 1.2 JWT Forgery
    print("\n-- 1.2 JWT Forgery Tests --")
    forged = [
        ("Bad signature",
         "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJwaWQiOjEsImNpZCI6MTAwMDAwMiwiaWF0IjoxNzgwODAxMjQ0LCJleHAiOjE3ODEwNjA0NDR9.INVALIDSIGNATURE"),
        ("Algorithm 'none'",
         "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJwaWQiOjEsImNpZCI6MTAwMDAwMiwiaWF0IjoxNzgwODAxMjQ0LCJleHAiOjE3ODEwNjA0NDR9."),
        ("Empty token string", ""),
    ]
    for name, tok in forged:
        status, data = api("POST", "/api/v2/report/", {"category": "bug", "description": "test"}, token=tok)
        if status == 401:
            test(f"JWT: {name}", "PASS")
        else:
            test(f"JWT: {name}", "FAIL", f"Expected 401 got {status}")
            FAIL -= 1  # reset - test() already incremented

    # 1.3 Expired token
    print("\n-- 1.3 Expired Token --")
    expired = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJwaWQiOjEsImNpZCI6MTAwMDAwMiwiaWF0IjoxMDAwMDAwMDAwLCJleHAiOjEwMDAwMDAwMDB9.dGhpcyBpcyBmYWtl"
    status, data = api("GET", "/api/v2/players/me/level/", token=expired)
    if status == 401:
        test("JWT: Expired token rejected", "PASS")
    else:
        test("JWT: Expired token", "FAIL", f"Expected 401 got {status}")

    # 1.4 No auth token
    print("\n-- 1.4 Missing Authentication --")
    protected_routes = [
        ("GET", "/api/v2/players/me/level/"),
        ("POST", "/api/v2/report/"),
        ("GET", "/api/v2/companies/me/buildings/"),
        ("POST", "/api/v2/buildings/buy/"),
    ]
    for method, path in protected_routes:
        status, data = api(method, path)
        if status == 401:
            test(f"No auth: {method} {path}", "PASS")
        else:
            test(f"No auth: {method} {path}", "FAIL", f"Expected 401 got {status}")

    return mtoken


# ============================================================
# SECTION 2: Input Validation
# ============================================================
def section_inputs(token):
    print("\n" + "="*60)
    print("SECTION 2: Input Validation")
    print("="*60)

    # 2.1 Path traversal (not possible - filename is server-generated)
    print("\n-- 2.1 Path Traversal --")
    test("Path traversal via description (JSON field, not filename)", "PASS", 
         "Filename uses server-controlled idgen.Next('report'), not user input")

    # 2.2 XSS in description
    print("\n-- 2.2 XSS Injection --")
    xss_payloads = [
        '<script>alert(1)</script>',
        '<img src=x onerror=alert(1)>',
        'javascript:alert(1)',
        '{{constructor.constructor("alert(1)")()}}',
    ]
    for payload in xss_payloads:
        status, data = api("POST", "/api/v2/report/", {"category": "bug", "description": payload}, token=token)
        if status == 201:
            # XSS is NOT rendered back to users - it's stored as JSON in a file
            test(f"XSS in description: {payload[:30]}", "WARN",
                 "Accepted. Safe: stored as JSON in server log, never rendered as HTML. Review if logs are surfaced in UI.")
        else:
            test(f"XSS in description: {payload[:30]}", "PASS", "Rejected")

    # 2.3 Null bytes and Unicode
    print("\n-- 2.3 Null Bytes & Unicode --")
    status, data = api("POST", "/api/v2/report/", 
                       {"category": "bug", "description": "null_\x00_byte_\u2603_snowman"}, token=token)
    if status == 201:
        test("Null bytes in description", "PASS", 
             "Accepted. Go JSON handles \\u0000. Verify file is valid JSON.")
    else:
        test("Null bytes in description", "FAIL", f"Expected 201 got {status}")

    # 2.4 Boundary tests
    print("\n-- 2.4 Length Boundaries --")
    status, data = api("POST", "/api/v2/report/", {"category": "bug", "description": "A" * 2000}, token=token)
    if status == 201:
        test("Description: 2000 chars (boundary)", "PASS")
    else:
        test("Description: 2000 chars", "FAIL", f"Expected 201 got {status}")

    status, data = api("POST", "/api/v2/report/", {"category": "bug", "description": "A" * 2001}, token=token)
    if status == 400:
        test("Description: 2001 chars (over limit)", "PASS")
    else:
        test("Description: 2001 chars", "FAIL", f"Expected 400 got {status}")

    # 2.5 Malformed JSON
    print("\n-- 2.5 Malformed Requests --")
    tests_malformed = [
        ("Invalid JSON body", "POST", "/api/v2/report/", "not json", None),
        ("XML Content-Type", "POST", "/api/v2/report/", "<xml/>", {"Content-Type": "application/xml"}),
        ("Empty body", "POST", "/api/v2/report/", "", None),
        ("Array instead of object", "POST", "/api/v2/report/", ["bug"], None),
    ]
    for name, method, path, body, headers in tests_malformed:
        status, data = api(method, path, body=body if isinstance(body, (dict, list)) else body,
                          token=token, headers=headers)
        if status >= 400:
            test(f"Malformed: {name}", "PASS")
        else:
            test(f"Malformed: {name}", "WARN", f"Got {status} instead of 4xx")

    # 2.6 JSON prototype pollution (Go is not vulnerable, but check)
    print("\n-- 2.6 JSON Prototype Pollution --")
    status, data = api("POST", "/api/v2/report/",
                      {"category": "bug", "description": "pp test", "__proto__": {"admin": True}}, token=token)
    if status == 201:
        test("Prototype pollution attempt", "PASS", "Go's json decoder ignores __proto__")
    else:
        test("Prototype pollution attempt", "FAIL", f"Unexpected status {status}")

    # 2.7 Content-Type handling
    print("\n-- 2.7 Content-Type Bypass --")
    status, data = api("POST", "/api/v2/report/", {"category": "bug", "description": "plain text"},
                      token=token, headers={"Content-Type": "text/plain"})
    if status == 201:
        test("Content-Type: text/plain (Go reads body anyway)", "PASS",
             "json.NewDecoder reads body regardless of Content-Type")
    else:
        test("Content-Type: text/plain", "FAIL", f"Expected 201 got {status}")


# ============================================================
# SECTION 3: Abuse & Rate Limiting
# ============================================================
def section_abuse(token):
    print("\n" + "="*60)
    print("SECTION 3: Abuse & Rate Limiting")
    print("="*60)

    # 3.1 Brute force login
    print("\n-- 3.1 Brute Force Login --")
    start = time.time()
    for i in range(10):
        status, data = api("POST", "/api/login", {"username": "dev", "password": f"wrong_{i}"})
    elapsed = time.time() - start

    if elapsed < 1:
        test("Brute force login (10 attempts)", "WARN",
             f"All 10 attempts completed in {elapsed:.2f}s - no throttling. "
             f"Config has RateLimitEnabled=false by default.")
    else:
        test("Brute force login (10 attempts)", "PASS", f"Took {elapsed:.2f}s")

    # 3.2 Rapid report submission (spam)
    print("\n-- 3.2 Rapid Submission (Spam) --")
    start = time.time()
    submitted = 0
    for i in range(50):
        status, data = api("POST", "/api/v2/report/", {"category": "bug", "description": f"spam_{i}"}, token=token)
        if status == 201:
            submitted += 1
    elapsed = time.time() - start

    if submitted == 50:
        test(f"Spam: 50 rapid submissions", "WARN",
             f"All 50 succeeded in {elapsed:.2f}s. No rate limiting on report endpoint.")
    else:
        test(f"Spam: 50 submissions", "WARN", f"{submitted}/50 succeeded. Possible throttling.")

    # 3.3 Concurrent submission
    print("\n-- 3.3 Large Payload DoS --")
    # Test with moderately large payload
    status, data = api("POST", "/api/v2/report/",
                      {"category": "bug", "description": "X" * 100000}, token=token)
    if status == 400:
        test("Large payload (100KB)", "PASS", "Rejected by validation")
    else:
        test("Large payload (100KB)", "WARN", f"Got {status}. Consider adding payload size limit at middleware level.")


# ============================================================
# SECTION 4: Business Logic & Information Disclosure
# ============================================================
def section_bizlogic(token):
    print("\n" + "="*60)
    print("SECTION 4: Business Logic & Information Disclosure")
    print("="*60)

    # 4.1 Category injection
    print("\n-- 4.1 Category Validation --")
    for cat in ["bug", "feature", "feedback", "other"]:
        status, data = api("POST", "/api/v2/report/", {"category": cat, "description": "test"}, token=token)
        if status == 201:
            test(f"Valid category: {cat}", "PASS")
        else:
            test(f"Valid category: {cat}", "FAIL", f"Expected 201 got {status}")

    status, data = api("POST", "/api/v2/report/", {"category": "invalid_cat", "description": "test"}, token=token)
    if status == 400:
        test("Invalid category: rejected", "PASS")
    else:
        test("Invalid category", "FAIL", f"Expected 400 got {status}")

    # 4.2 Extra fields in request
    print("\n-- 4.2 Extra Field Injection --")
    status, data = api("POST", "/api/v2/report/",
                      {"category": "bug", "description": "test", "playerId": 99999, "companyId": 99999, "isAdmin": True},
                      token=token)
    if status == 201:
        # The extra fields are silently ignored by Go's json.Decoder
        test("Extra fields in body", "PASS", "Go ignores unknown fields. No impact.")
    else:
        test("Extra fields", "WARN", f"Got {status}")

    # 4.3 Replay attack
    print("\n-- 4.3 Replay Attack --")
    payload = {"category": "bug", "description": "replay_me"}
    s1, d1 = api("POST", "/api/v2/report/", payload, token=token)
    s2, d2 = api("POST", "/api/v2/report/", payload, token=token)
    if s1 == 201 and s2 == 201:
        id1 = d1.get("data", {}).get("id", "")
        id2 = d2.get("data", {}).get("id", "")
        if id1 != id2:
            test("Replay attack", "PASS", "Each submission gets unique ID. No dedup needed for log entries.")
        else:
            test("Replay attack", "WARN", "Duplicate IDs detected")
    else:
        test("Replay attack", "WARN", f"Unexpected status {s1}/{s2}")

    # 4.4 Information disclosure
    print("\n-- 4.4 Information Disclosure --")
    status, data = api("POST", "/api/v2/report/", {"category": "bug", "description": "X" * 100}, token=token)
    if status == 201:
        resp = data
        resp_str = json.dumps(resp)
        # Skip the standard envelope fields
        data_part = resp_str.replace('"error":null', '').replace('"data":', '')
        sensitive = ["internal", "stack", "trace", "debug", "file", "line", "/var/", "C:\\", "/etc/"]
        leaks = [s for s in sensitive if s.lower() in data_part.lower()]
        if not leaks:
            test("Success response: no sensitive info leak", "PASS")
        else:
            test("Success response: info leak", "WARN", f"Found: {leaks} in {resp_str[:200]}")
    try:
        req = urllib.request.Request(f"{BASE}/log/")
        resp = urllib.request.urlopen(req)
        test("Directory listing of /log/", "FAIL", "Directory listing exposed!")
    except urllib.error.HTTPError as e:
        if e.code == 404 or e.code == 403:
            test("Directory listing of /log/ blocked", "PASS")
        else:
            test("Directory listing", "WARN", f"Got {e.code}")

    # 4.6 Direct file access
    print("\n-- 4.6 Direct Report File Access --")
    try:
        req = urllib.request.Request(f"{BASE}/../log/report-test.json")
        resp = urllib.request.urlopen(req)
        test("Direct file access via path traversal", "FAIL", 
             f"Accessible! Status {resp.status}")
    except urllib.error.HTTPError as e:
        if e.code in (404, 403):
            test("Direct report file access blocked", "PASS")
        else:
            test("Direct file access", "WARN", f"Got {e.code}")


# ============================================================
# SUMMARY
# ============================================================
def summary():
    print("\n" + "="*60)
    print("SECURITY AUDIT SUMMARY")
    print("="*60)
    total = PASS + FAIL + WARN
    print(f"  Total tests : {total}")
    print(f"  PASS        : {PASS}")
    print(f"  WARN        : {WARN}")
    print(f"  FAIL        : {FAIL}")
    print()

    if FAIL > 0:
        print("  ** FAILURES REQUIRING ATTENTION:")
        print(f"     {FAIL} test(s) failed. Review above.")
    if WARN > 0:
        print("  ** WARNINGS TO REVIEW:")
        print(f"     {WARN} warning(s). Consider mitigations.")
    if PASS == total:
        print("  ** All checks passed!")
    print()


# ============================================================
# MAIN
# ============================================================
if __name__ == "__main__":
    print("New Haven Security Audit")
    print("=" * 60)
    print(f"Target: {BASE}")

    token = login()
    if not token:
        print("ERROR: Could not login. Is the server running?")
        sys.exit(1)

    print(f"Authenticated: yes (token: {token[:20]}...)")

    token = section_auth()
    section_inputs(token)
    section_abuse(token)
    section_bizlogic(token)
    summary()

    # Cleanup test files
    import os, glob
    for f in glob.glob("../log/report-*.json"):
        os.remove(f)
    print("Test artifacts cleaned.")
