package handlers

import "testing"

func TestIsBlockedHost(t *testing.T) {
	blocked := []string{
		"localhost",
		"127.0.0.1",
		"127.0.0.1:8080",
		"10.0.0.1",
		"192.168.1.3",
		"172.16.0.1",
		"[::1]",
		"",
	}
	for _, h := range blocked {
		if !isBlockedHost(h) {
			t.Errorf("isBlockedHost(%q) = false, 期望拒绝", h)
		}
	}
}
