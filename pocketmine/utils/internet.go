package utils

import (
	"io"
	"net"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"
)

var (
	internetMu  sync.Mutex
	OnlineMode  = true
	cachedIP    string
	hasCachedIP bool
)

// GetIP lazily gets the external IP using an external service and caches the result.
func GetIP(force bool) (string, bool) {
	internetMu.Lock()
	online := OnlineMode
	cached, has := cachedIP, hasCachedIP
	internetMu.Unlock()
	if !online {
		return "", false
	}
	if has && !force {
		return cached, true
	}

	remember := func(ip string) (string, bool) {
		internetMu.Lock()
		cachedIP, hasCachedIP = ip, true
		internetMu.Unlock()
		return ip, true
	}

	if res, err := GetURL("http://api.ipify.org/", 10, nil); err == nil {
		return remember(res.Body)
	}
	if res, err := GetURL("http://checkip.dyndns.org/", 10, nil); err == nil {
		if m := regexp.MustCompile(`Current IP Address: ([0-9a-fA-F:.]*)`).FindStringSubmatch(strings.TrimSpace(stripTags(res.Body))); m != nil {
			return remember(m[1])
		}
	}
	if res, err := GetURL("http://www.checkip.org/", 10, nil); err == nil {
		if m := regexp.MustCompile(`">([0-9a-fA-F:.]*)</span>`).FindStringSubmatch(res.Body); m != nil {
			return remember(m[1])
		}
	}
	if res, err := GetURL("http://checkmyip.org/", 10, nil); err == nil {
		if m := regexp.MustCompile(`Your IP address is ([0-9a-fA-F:.]*)`).FindStringSubmatch(res.Body); m != nil {
			return remember(m[1])
		}
	}
	if res, err := GetURL("http://ifconfig.me/ip", 10, nil); err == nil {
		if addr := strings.TrimSpace(res.Body); addr != "" {
			return remember(addr)
		}
	}
	return "", false
}

var htmlTagPattern = regexp.MustCompile(`<[^>]*>`)

func stripTags(s string) string {
	return htmlTagPattern.ReplaceAllString(s, "")
}

// GetInternalIP returns the machine's internal network IP address, by opening a UDP "connection"
// (no packets are actually sent) to a well-known external address and reading back the local
// address the OS routed it through — the same trick the PHP original uses via raw sockets.
func GetInternalIP() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:65534")
	if err != nil {
		return "", &InternetException{Message: "Failed to get internal IP: " + err.Error()}
	}
	defer conn.Close()
	return conn.LocalAddr().(*net.UDPAddr).IP.String(), nil
}

var httpClient = &http.Client{}

// GetURL GETs a URL. NOTE: this is a blocking operation and can take a significant amount of
// time; avoid calling it from latency-sensitive code paths.
func GetURL(page string, timeoutSeconds float64, extraHeaders map[string]string) (*InternetRequestResult, error) {
	return simpleRequest(http.MethodGet, page, nil, timeoutSeconds, extraHeaders)
}

// PostURL POSTs data to a URL. NOTE: this is a blocking operation and can take a significant
// amount of time; avoid calling it from latency-sensitive code paths.
func PostURL(page string, body io.Reader, timeoutSeconds float64, extraHeaders map[string]string) (*InternetRequestResult, error) {
	return simpleRequest(http.MethodPost, page, body, timeoutSeconds, extraHeaders)
}

// simpleRequest is a port of Internet::simpleCurl().
//
// The PHP original disables TLS certificate verification (CURLOPT_SSL_VERIFYPEER/VERIFYHOST
// set to false/2-meaning-off) for every request. That's a real security weakness rather than a
// deliberate requirement of the protocol being spoken, so this port deliberately does NOT carry
// it over — it uses Go's default, verified TLS configuration instead.
func simpleRequest(method, page string, body io.Reader, timeoutSeconds float64, extraHeaders map[string]string) (*InternetRequestResult, error) {
	if !OnlineMode {
		return nil, &InternetException{Message: "Cannot execute web request while offline"}
	}

	req, err := http.NewRequest(method, page, body)
	if err != nil {
		return nil, &InternetException{Message: "Unable to create request: " + err.Error()}
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1; WOW64; rv:12.0) Gecko/20100101 Firefox/12.0 PocketMine-Go")
	for k, v := range extraHeaders {
		req.Header.Set(k, v)
	}

	client := *httpClient
	client.Timeout = time.Duration(timeoutSeconds * float64(time.Second))

	resp, err := client.Do(req)
	if err != nil {
		return nil, &InternetException{Message: err.Error()}
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &InternetException{Message: err.Error()}
	}

	headers := make([]map[string]string, 1)
	headers[0] = make(map[string]string, len(resp.Header))
	for k, v := range resp.Header {
		if len(v) > 0 {
			headers[0][strings.ToLower(k)] = strings.TrimSpace(v[0])
		}
	}

	return &InternetRequestResult{
		Headers: headers,
		Body:    string(bodyBytes),
		Code:    resp.StatusCode,
	}, nil
}
