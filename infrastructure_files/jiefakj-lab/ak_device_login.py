#!/usr/bin/env python3
"""模拟用户在浏览器完成 Authentik 登录 + RFC8628 设备码授权。

用法: ak_device_login.py <username> <password> <user_code>
仅用标准库; 基址来自 AUTH_BASE_URL(默认 https://auth.example.com)。
"""
import http.cookiejar
import json
import os
import ssl
import sys
import urllib.parse
import urllib.request

BASE = os.environ.get("AUTH_BASE_URL", "https://auth.example.com")
CTX = ssl.create_default_context()


def main() -> None:
    username, password, user_code = sys.argv[1], sys.argv[2], sys.argv[3]
    cj = http.cookiejar.CookieJar()
    opener = urllib.request.build_opener(
        urllib.request.HTTPCookieProcessor(cj),
        urllib.request.HTTPSHandler(context=CTX),
    )
    opener.addheaders = [("User-Agent", "ak-device-login-helper")]

    def api(flow_slug: str, payload=None, query: str = ""):
        url = f"{BASE}/api/v3/flows/executor/{flow_slug}/?query={urllib.parse.quote(query)}"
        if payload is None:
            req = urllib.request.Request(url)
        else:
            req = urllib.request.Request(
                url,
                data=json.dumps(payload).encode(),
                headers={"Content-Type": "application/json"},
            )
        with opener.open(req, timeout=20) as r:
            return json.loads(r.read())

    def run_flow(slug: str, query: str = "") -> str:
        ch = api(slug, None, query)
        for _ in range(12):
            comp = ch.get("component", "")
            errs = ch.get("response_errors") or {}
            print(f"    [challenge] {comp} type={ch.get('type')} errors={errs if errs else ''}", file=sys.stderr)
            if comp == "ak-stage-identification":
                payload = {"uid_field": username, "component": comp}
                if ch.get("password_fields"):
                    payload["password"] = password
                ch = api(slug, payload, query)
            elif comp == "ak-stage-password":
                ch = api(slug, {"password": password, "component": comp}, query)
            elif comp == "ak-provider-oauth2-device-code":
                ch = api(slug, {"code": user_code, "component": comp}, query)
            elif comp == "ak-stage-consent":
                ch = api(slug, {"component": comp}, query)
            elif comp == "ak-provider-oauth2-device-code-finish" or ch.get("type") == "empty":
                return "device-done"
            elif comp == "xak-flow-redirect":
                return ch.get("to", "redirect")
            elif comp == "ak-stage-access-denied":
                raise SystemExit(f"access denied: {json.dumps(ch)[:300]}")
            else:
                raise SystemExit(f"unhandled challenge: {json.dumps(ch)[:500]}")
        raise SystemExit("flow did not converge")

    print("[1] authenticating as", username, "->", run_flow("default-authentication-flow"))

    with opener.open(f"{BASE}/device?code={urllib.parse.quote(user_code)}", timeout=20) as r:
        final = r.geturl()
    print("[2] device init landed on:", final)
    if "/if/flow/" in final:
        slug = final.split("/if/flow/")[1].split("/")[0]
        query = urllib.parse.urlsplit(final).query
        print("[3] executing flow", slug, "->", run_flow(slug, query))
    else:
        print("[3] no flow to execute (already done?)")


if __name__ == "__main__":
    main()
