package middleware

import "testing"

func TestRateLimitKeyNormalizationAndIsolation(t *testing.T) {
	key := rateLimitKey("app", "login", "account", " 202610270001 ")
	if key != rateLimitKey("app", "login", "account", "202610270001") {
		t.Fatal("account whitespace must match login normalization")
	}
	if key != "leave:app:login:account:202610270001" {
		t.Fatal(key)
	}
	for _, other := range []string{
		rateLimitKey("admin", "login", "account", "202610270001"),
		rateLimitKey("app", "register", "account", "202610270001"),
		rateLimitKey("app", "login", "account", "202610270002"),
	} {
		if key == other {
			t.Fatal("key collision")
		}
	}
	if rateLimitKey("app", "general", "ip", "::ffff:192.0.2.1") != rateLimitKey("app", "general", "ip", "192.0.2.1") {
		t.Fatal("mapped IPv4 must share quota")
	}
	if rateLimitKey("admin", "login", "account", "Admin") == rateLimitKey("admin", "login", "account", "admin") {
		t.Fatal("case must be preserved")
	}
}
