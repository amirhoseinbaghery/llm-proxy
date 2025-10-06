package auth

import (
	"context"
)

func GetClaimsFromContext(ctx context.Context) (*Claims, bool) {
	claims, ok := ctx.Value("user_claims").(*Claims)
	return claims, ok
}
