package cookies

import (
	"context"
	"net/http"

	"github.com/justinsb/kweb/components"
)

func NewCookiesComponent() components.Component {
	return &CookiesComponent{}
}

type CookiesComponent struct {
}

func (c *CookiesComponent) RegisterHandlers(s *components.Server, mux *http.ServeMux) error {
	return nil
}

func (c *CookiesComponent) Key() string {
	return "cookies"
}

func (c *CookiesComponent) ScopeValues() any {
	return nil
}

var contextKeyResponseCookies = &responseCookies{}

// getResponseCookies returns the ResponseCookies object used to add cookies to the response.
func getResponseCookies(ctx context.Context) *responseCookies {
	cookies := ctx.Value(contextKeyResponseCookies)
	return cookies.(*responseCookies)
}

var _ components.RequestFilter = &CookiesComponent{}

func (c *CookiesComponent) ProcessRequest(ctx context.Context, req *components.Request, next components.RequestFilterChain) (components.Response, error) {
	responseCookies := &responseCookies{}

	ctx = context.WithValue(ctx, contextKeyResponseCookies, responseCookies)

	response, err := next(ctx, req)
	if err != nil {
		return nil, err
	}

	if len(responseCookies.setCookies) != 0 {
		addCookiesResponse := &addCookiesResponse{
			responseCookies: responseCookies,
			inner:           response,
		}
		response = addCookiesResponse
	}

	return response, err
}

// addCookiesResponse wraps a Response but adds Set-Cookie headers to the response
type addCookiesResponse struct {
	responseCookies *responseCookies
	inner           components.Response
}

func (r *addCookiesResponse) WriteTo(ctx context.Context, w http.ResponseWriter) {
	for i := range r.responseCookies.setCookies {
		cookie := &r.responseCookies.setCookies[i]
		http.SetCookie(w, cookie)
	}
	r.inner.WriteTo(ctx, w)
}
