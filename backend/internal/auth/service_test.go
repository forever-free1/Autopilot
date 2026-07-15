package auth

import "testing"

func TestPasswordAndToken(t *testing.T) {
	hash, err := HashPassword("autopilot123")
	if err != nil || !CheckPassword(hash, "autopilot123") || CheckPassword(hash, "wrong") {
		t.Fatal("密码哈希校验未按预期工作")
	}
	token, err := CreateToken(7, "test-secret")
	if err != nil {
		t.Fatalf("创建令牌失败: %v", err)
	}
	userID, err := ParseToken(token, "test-secret")
	if err != nil || userID != 7 {
		t.Fatalf("解析令牌失败: userID=%d err=%v", userID, err)
	}
}

func TestPasswordLength(t *testing.T) {
	if _, err := HashPassword("short"); err == nil {
		t.Fatal("过短密码应被拒绝")
	}
}
