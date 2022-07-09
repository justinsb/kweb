package cookies

import (
	"context"
	"net/http"
)

type responseCookies struct {
	setCookies []http.Cookie
}

func (r *responseCookies) SetCookie(cookie http.Cookie) {
	r.setCookies = append(r.setCookies, cookie)
}

// SetCookie will add a cookie to the outgoing response
func SetCookie(ctx context.Context, cookie http.Cookie) {
	responseCookies := getResponseCookies(ctx)
	responseCookies.SetCookie(cookie)
}
