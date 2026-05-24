package dingtalk

import "net/url"

func parseQueryFromURL(rawURL string) (url.Values, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	return u.Query(), nil
}
