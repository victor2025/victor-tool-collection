package handlers

import "testing"

func TestVerifyPassword(t *testing.T) {
	hashed, err := HashPassword("secret123")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if !IsBcryptHash(hashed) {
		t.Fatalf("生成的哈希未识别为 bcrypt: %s", hashed)
	}

	// 正确密码 + bcrypt 存储
	if ok, upgrade := verifyPassword(hashed, "secret123"); !ok || upgrade {
		t.Fatalf("bcrypt 正确密码应通过且不需升级，得到 ok=%v upgrade=%v", ok, upgrade)
	}
	// 错误密码
	if ok, _ := verifyPassword(hashed, "secret124"); ok {
		t.Fatal("bcrypt 错误密码不应通过")
	}
	// 历史明文：通过且需升级
	if ok, upgrade := verifyPassword("secret123", "secret123"); !ok || !upgrade {
		t.Fatalf("明文应通过且需升级，得到 ok=%v upgrade=%v", ok, upgrade)
	}
	// 空密码
	if ok, _ := verifyPassword("", "x"); ok {
		t.Fatal("空存储值不应通过")
	}
	if ok, _ := verifyPassword("secret123", ""); ok {
		t.Fatal("空输入密码不应通过")
	}
}

func TestSanitizeTag(t *testing.T) {
	cases := []struct{ in, want string }{
		{"base64", "base64"},
		{"/base64/", "/base64/"},
		{"a b", "a-b"},
		{"<script>", "script"},
		{"ümlaut€", "mlaut"},
	}
	for _, c := range cases {
		if got := sanitizeTag(c.in, 64); got != c.want {
			t.Errorf("sanitizeTag(%q) = %q, want %q", c.in, got, c.want)
		}
	}
	if got := sanitizeTag("abcdefghij", 4); got != "abcd" {
		t.Errorf("截断失败: %q", got)
	}
}
