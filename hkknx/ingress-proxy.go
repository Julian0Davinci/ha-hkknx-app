// HKKNX Home Assistant Ingress path adapter.
// It proxies only Supervisor-originated requests to the unmodified HKKNX UI.
package main

import (
    "bytes"
    "errors"
    "fmt"
    "io"
    "log"
    "net"
    "net/http"
    "net/http/httputil"
    "net/url"
    "regexp"
    "strconv"
    "strings"
    "time"
)

const (
    listenAddress = ":8099"
    hkknxAddress  = "http://127.0.0.1:8080"
    maxRewriteBytes = 16 << 20
)

var (
    ingressPath = regexp.MustCompile(`^/api/hassio_ingress/[A-Za-z0-9_-]+/?$`)
    htmlAttributes = regexp.MustCompile(`(?i)(\b(?:href|src|action|poster|formaction|data-src|(?:data-)?hx-(?:get|post|put|patch|delete|push-url|replace-url)|(?:data-)?ws-connect)\s*=\s*["'])/([^/])`)
    htmlUnquoted = regexp.MustCompile(`(?i)(\b(?:href|src|action|poster|formaction|data-src|(?:data-)?hx-(?:get|post|put|patch|delete|push-url|replace-url)|(?:data-)?ws-connect)\s*=\s*)/([A-Za-z0-9])`)
    htmlBase = regexp.MustCompile(`(?i)<head(?:\s[^>]*)?>`)
    cssURLs = regexp.MustCompile(`(?i)(url\(\s*["']?)/([^/])`)
    // Script paths: rewriting source-literal root URLs is a best-effort bridge.
    // It does not alter dynamically constructed URLs or arbitrary API payloads.
    jsRootURLs = regexp.MustCompile(`(["'\x60])/([A-Za-z0-9._~%-])`)
    cookieRoot = regexp.MustCompile(`(?i)(;\s*Path\s*=\s*)/([;\s]|$)`)
)

func normalizeBase(value string) string {
    value = strings.TrimSpace(value)
    if !ingressPath.MatchString(value) {
        return ""
    }
    return strings.TrimRight(value, "/")
}

func fromSupervisor(r *http.Request) bool {
    host, _, err := net.SplitHostPort(r.RemoteAddr)
    if err != nil { return false }
    ip := net.ParseIP(host)
    return ip != nil && (ip.Equal(net.ParseIP("172.30.32.2")) || ip.IsLoopback())
}

func redirectLocation(original, base string) string {
    if base == "" || original == "" { return original }
    parsed, err := url.Parse(original)
    if err != nil { return original }
    if parsed.IsAbs() {
        // Only rewrite upstream-local absolute URLs, never third-party redirects.
        if parsed.Hostname() != "127.0.0.1" && parsed.Hostname() != "localhost" { return original }
        if parsed.Port() != "" && parsed.Port() != "8080" { return original }
    } else if strings.HasPrefix(original, "//") {
        return original
    }
    if !strings.HasPrefix(parsed.Path, "/") { return original }
    if parsed.Path == base || strings.HasPrefix(parsed.Path, base+"/") { return original }
    parsed.Scheme, parsed.Host = "", ""
    parsed.Path = base + parsed.Path
    return parsed.String()
}


const ingressAPIShim = `<script>
(function () {
  "use strict";
  // Keep API requests under the Home Assistant Ingress session URL.
  // The prefix is populated from Supervisor's X-Ingress-Path, not hardcoded.
  var ingressBase = "__INGRESS_BASE__";
  function ingressURL(input) {
    var url;
    try { url = new URL(input, window.location.href); }
    catch (_) { return input; }
    if (url.origin !== window.location.origin) return input;
    if (!url.pathname.startsWith("/api/") && !url.pathname.startsWith("/projects/")) return input;
    if (url.pathname.startsWith(ingressBase + "/")) return input;
    url.pathname = ingressBase + url.pathname;
    return url.href;
  }
  // htmx's ws extension constructs its own WebSocket URL. Redirect it to
  // the same authenticated Ingress session, preserving ws/wss and the query.
  function ingressWebSocketURL(input) {
    var url;
    try { url = new URL(String(input), window.location.href); }
    catch (_) { return input; }
    if (url.protocol !== "ws:" && url.protocol !== "wss:") return input;
    if (url.host !== window.location.host) return input;
    if (url.pathname.startsWith(ingressBase + "/")) return input;
    if (!url.pathname.startsWith("/projects/") && !url.pathname.startsWith("/api/")) return input;
    url.pathname = ingressBase + url.pathname;
    return url.href;
  }
  if (typeof window.WebSocket === "function") {
    var NativeWebSocket = window.WebSocket;
    function IngressWebSocket(url, protocols) {
      var destination = ingressWebSocketURL(url);
      if (arguments.length === 1) return new NativeWebSocket(destination);
      return new NativeWebSocket(destination, protocols);
    }
    IngressWebSocket.prototype = NativeWebSocket.prototype;
    Object.setPrototypeOf(IngressWebSocket, NativeWebSocket);
    window.WebSocket = IngressWebSocket;
  }
  if (typeof window.fetch === "function") {
    var nativeFetch = window.fetch;
    window.fetch = function (input, init) {
      if (typeof input === "string" || input instanceof URL) {
        return nativeFetch.call(this, ingressURL(String(input)), init);
      }
      if (input instanceof Request) {
        var rewritten = ingressURL(input.url);
        if (rewritten !== input.url) input = new Request(rewritten, input);
      }
      return nativeFetch.call(this, input, init);
    };
  }
  if (window.XMLHttpRequest && window.XMLHttpRequest.prototype) {
    var nativeOpen = window.XMLHttpRequest.prototype.open;
    window.XMLHttpRequest.prototype.open = function (method, url) {
      var args = Array.prototype.slice.call(arguments);
      if (typeof url === "string") args[1] = ingressURL(url);
      return nativeOpen.apply(this, args);
    };
  }
})();
</script>`

func rewriteText(original, contentType, base string) string {
    if base == "" { return original }
    lowerType := strings.ToLower(contentType)
    content := original
    switch {
    case strings.Contains(lowerType, "text/html"):
        // Rewrite first, so the injected <base> is not prefixed twice.
        content = htmlAttributes.ReplaceAllString(content, "${1}"+base+"/${2}")
        content = htmlUnquoted.ReplaceAllString(content, "${1}"+base+"/${2}")
        // Relative URLs also need a stable base after the /projects/... redirect.
        if !strings.Contains(strings.ToLower(content), "<base ") {
            content = htmlBase.ReplaceAllStringFunc(content, func(tag string) string {
                return tag + `<base href="` + base + `/">` + strings.ReplaceAll(ingressAPIShim, "__INGRESS_BASE__", base)
            })
        }
    case strings.Contains(lowerType, "text/css"):
        content = cssURLs.ReplaceAllString(content, "${1}"+base+"/${2}")
    case strings.Contains(lowerType, "javascript") || strings.Contains(lowerType, "ecmascript"):
        content = jsRootURLs.ReplaceAllString(content, "${1}"+base+"/${2}")
    }
    return content
}

func rewriteResponse(resp *http.Response) error {
    base := normalizeBase(resp.Request.Header.Get("X-Ingress-Path"))
    if base == "" { return errors.New("missing or invalid X-Ingress-Path") }

    if original := resp.Header.Get("Location"); original != "" {
        resp.Header.Set("Location", redirectLocation(original, base))
        resp.Header.Set("Cache-Control", "private, no-store")
    }
    cookies := resp.Header.Values("Set-Cookie")
    if len(cookies) > 0 {
        resp.Header.Del("Set-Cookie")
        for _, c := range cookies {
            // Path=/ is unsuitable for a page under the Ingress URL.
            // Also handles other root-scoped attributes without changing values.
            fixed := cookieRoot.ReplaceAllString(c, "${1}"+base+"/${2}")
            resp.Header.Add("Set-Cookie", fixed)
        }
    }
    ct := strings.ToLower(resp.Header.Get("Content-Type"))
    rewrite := strings.Contains(ct, "text/html") || strings.Contains(ct, "text/css") || strings.Contains(ct, "javascript") || strings.Contains(ct, "ecmascript")
    if !rewrite || resp.Body == nil || resp.StatusCode == http.StatusNotModified { return nil }
    if resp.Header.Get("Content-Encoding") != "" { return nil } // no unsafe edits to compressed bytes

    data, err := io.ReadAll(io.LimitReader(resp.Body, maxRewriteBytes+1))
    if err != nil { return fmt.Errorf("read HKKNX response: %w", err) }
    if len(data) > maxRewriteBytes {
        resp.Body = io.NopCloser(io.MultiReader(bytes.NewReader(data), resp.Body))
        return nil
    }
    resp.Body.Close()
    result := []byte(rewriteText(string(data), ct, base))
    resp.Body = io.NopCloser(bytes.NewReader(result))
    resp.ContentLength = int64(len(result))
    resp.Header.Set("Content-Length", strconv.Itoa(len(result)))
    resp.Header.Del("ETag")
    resp.Header.Del("Content-MD5")
    resp.Header.Set("Cache-Control", "private, no-store")
    return nil
}

func newHandler() http.Handler {
    upstream, _ := url.Parse(hkknxAddress)
    proxy := httputil.NewSingleHostReverseProxy(upstream)
    originalDirector := proxy.Director
    proxy.Director = func(req *http.Request) {
        originalDirector(req)
        req.Host = upstream.Host
        // Avoid upstream-compressed content when rewriting HTML, CSS and JS.
        req.Header.Del("Accept-Encoding")
    }
    proxy.ModifyResponse = rewriteResponse
    proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
        log.Printf("Ingress proxy error: %v", err)
        http.Error(w, "HKKNX ingress proxy error", http.StatusBadGateway)
    }
    return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
        if !fromSupervisor(req) { http.Error(w, "Forbidden", http.StatusForbidden); return }
        if normalizeBase(req.Header.Get("X-Ingress-Path")) == "" {
            http.Error(w, "Missing Ingress base path", http.StatusBadRequest)
            return
        }
        proxy.ServeHTTP(w, req)
    })
}

func main() {
    srv := &http.Server{Addr: listenAddress, Handler: newHandler(), ReadHeaderTimeout: 10*time.Second}
    log.Printf("HKKNX Ingress proxy listening on %s, upstream %s", listenAddress, hkknxAddress)
    log.Fatal(srv.ListenAndServe())
}
