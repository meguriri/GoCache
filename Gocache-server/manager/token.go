package manager

import (
	"context"
	"math/rand"
	"strconv"
	"time"
)

// Token token认证
type Token struct {
	Secret string
}

// GetRequestMetadata 获取当前请求认证所需的元数据（metadata）
func (t *Token) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	// 设置一个种子
	rand.Seed(time.Now().UnixNano())
	// Intn返回一个取值范围在[0,n)的伪随机int值
	num := rand.Intn(100) + 1 // 随机1-100
	rangeSeed := strconv.Itoa(num)
	return map[string]string{"secret": t.Secret, "range_seed": rangeSeed}, nil
}

// RequireTransportSecurity 是否需要基于 TLS 认证进行安全传输,返回false不进行TLS验证
func (t *Token) RequireTransportSecurity() bool {
	return false
}
