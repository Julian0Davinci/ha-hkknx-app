package main

import (
    "io"
    "net/http"
    "net/url"
    "strings"
    "testing"
)

const base = "/api/hassio_ingress/MyToken_-1"

func TestRedirectLocation(t *testing.T) {
    for _, tc := range []struct{ in,out string }{
        {"/projects/77c?", base+"/projects/77c?"},
        {"/",base+"/"},
        {"/assets/app.js",base+"/assets/app.js"},
        {"https://example.com/projects/77c", "https://example.com/projects/77c"},
        {"//example.com/path", "//example.com/path"},
        {base+"/projects",base+"/projects"},
    } {
        if got:=redirectLocation(tc.in,base); got!=tc.out { t.Errorf("%q: got %q want %q",tc.in,got,tc.out) }
    }
}

func TestHTMLRewrites(t *testing.T) {
    in := `<html><head><title>HKKNX</title></head><body><a href="/projects/a">Project</a><script src="/static/main.js"></script><img src="https://example.com/img.png"></body></html>`
    out:=rewriteText(in,"text/html",base)
    for _, expected := range []string{`<base href="`+base+`/"> `,`href="`+base+`/projects/a"`,`src="`+base+`/static/main.js"`,`src="https://example.com/img.png"`} {
        if !strings.Contains(out,strings.TrimSpace(expected)) {t.Errorf("missing %s\n%s",expected,out)}
    }
}

func TestJSAndCSS(t *testing.T) {
    js := "var regex = /foo/gi;"
    if got := rewriteText(js, "application/javascript", base); got != js { t.Errorf("JavaScript modified: %s", got) }
    css := "body {color: red;}"
    if got := rewriteText(css, "text/css", base); got != css { t.Errorf("CSS unexpectedly modified: %s", got) }
}
func TestRewriteResponse(t *testing.T) {
    req,_:=http.NewRequest("GET","http://localhost/",nil)
    req.Header.Set("X-Ingress-Path",base)
    resp:=&http.Response{StatusCode:303,Header:make(http.Header),Body:io.NopCloser(strings.NewReader("")),Request:req}
    resp.Header.Set("Location","/projects/a?")
    resp.Header.Add("Set-Cookie","hkknx=abc; Path=/; HttpOnly")
    if err:=rewriteResponse(resp);err!=nil {t.Fatal(err)}
    if got:=resp.Header.Get("Location");got!=base+"/projects/a?" {t.Errorf("location %q",got)}
    if got:=resp.Header.Get("Set-Cookie");!strings.Contains(got,"Path="+base+"/") {t.Errorf("cookie %q",got)}
}

func TestInvalidBase(t *testing.T) {
    if normalizeBase("/fake/x")!="" {t.Fatal("unexpected accepted base")}
    if normalizeBase(base)!=base {t.Fatal("valid base rejected")}
    if _,err:=url.Parse(base);err!=nil {t.Fatal(err)}
}

func TestIngressAPIShimInjection(t *testing.T) {
    in := `<html><head><script src="/static/main.js"></script></head><body></body></html>`
    out := rewriteText(in, "text/html", base)
    for _, expected := range []string{
        `var ingressBase = "` + base + `";`,
        `window.fetch = function`,
        `window.XMLHttpRequest.prototype.open = function`,
        `src="` + base + `/static/main.js"`,
    } {
        if !strings.Contains(out, expected) {
            t.Errorf("missing injected ingress adapter: %s", expected)
        }
    }
    if strings.Index(out, "window.fetch = function") > strings.Index(out, "static/main.js") {
        t.Fatal("API shim was injected after HKKNX scripts")
    }
}

func TestHTMXRewrites(t *testing.T) {
    in := `<form hx-post="/api/validate-credentials?pid=test"><button hx-get="/projects/abc/hx-chart-container?chartId=x" hx-target="#chart">Open</button><div data-hx-get='/projects/abc/hx-status'></div><a href="/projects/abc">Go</a></form>`
    out := rewriteText(in, "text/html", base)
    for _, expected := range []string{
        `hx-post="` + base + `/api/validate-credentials?pid=test"`,
        `hx-get="` + base + `/projects/abc/hx-chart-container?chartId=x"`,
        `data-hx-get='` + base + `/projects/abc/hx-status'`,
        `href="` + base + `/projects/abc"`,
    } {
        if !strings.Contains(out, expected) { t.Errorf("missing %q in %q", expected, out) }
    }
    if !strings.Contains(out, `!url.pathname.startsWith("/projects/")`) {
        t.Errorf("client-side shim must also route dynamically generated HTMX project requests")
    }
}

func TestWebSocketMarkupRewrites(t *testing.T) {
    in := `<html><head></head><body><div ws-connect="/projects/abc/hx-ws-accessory-states" data-ws-connect='/projects/abc/hx-ws-accessory-states'></div></body></html>`
    out := rewriteText(in, "text/html", base)
    for _, want := range []string{
        `ws-connect="` + base + `/projects/abc/hx-ws-accessory-states"`,
        `data-ws-connect='` + base + `/projects/abc/hx-ws-accessory-states'`,
        `window.WebSocket = IngressWebSocket;`,
        `url.host !== window.location.host`,
        `url.pathname.startsWith(ingressBase + "/")`,
    } {
        if !strings.Contains(out, want) {
            t.Errorf("missing WebSocket adapter element %q", want)
        }
    }
    if strings.Index(out, "window.WebSocket = IngressWebSocket") > strings.Index(out, "ws-connect=") {
        t.Fatal("WebSocket adapter must load before HTML executes its HTMX websocket extension")
    }
}

func TestHTMXWebSocketScheme(t *testing.T) {
    in := `<div ws-connect="ws:/projects/x/hx-ws-accessory-states" data-ws-connect="wss:/projects/x/status"></div>`
    out := rewriteText(in, "text/html", base)
    if !strings.Contains(out, "ws:"+base+"/projects/x/hx-ws-accessory-states") { t.Errorf("missing ws prefix: %s",out) }
    if !strings.Contains(out, "wss:"+base+"/projects/x/status") { t.Errorf("missing wss prefix: %s",out) }
}

func TestWebSocketOriginNormalization(t *testing.T) {
    upstream, err := url.Parse("http://127.0.0.1:8080")
    if err != nil { t.Fatal(err) }
    ws, err := http.NewRequest("GET", "http://127.0.0.1:8080/projects/example/hx-ws-accessory-states", nil)
    if err != nil { t.Fatal(err) }
    ws.Header.Set("Upgrade", "websocket")
    ws.Header.Set("Origin", "http://192.168.1.116:8123")
    normalizeUpstreamWebSocketOrigin(ws, upstream)
    if got := ws.Header.Get("Origin"); got != "http://127.0.0.1:8080" { t.Errorf("origin = %q", got) }
    normal, err := http.NewRequest("GET", "http://127.0.0.1:8080/projects/example", nil)
    if err != nil { t.Fatal(err) }
    normal.Header.Set("Origin", "http://192.168.1.116:8123")
    normalizeUpstreamWebSocketOrigin(normal, upstream)
    if got := normal.Header.Get("Origin"); got != "http://192.168.1.116:8123" { t.Errorf("normal HTTP origin changed to %q", got) }
}

func TestAccessoryImageDynamicIngressShim(t *testing.T) {
    html := `<html><head></head><body><img src="/static/img/accessory/v2/active/humidity-sensor.png"></body></html>`
    output := rewriteText(html, "text/html", base)
    for _, want := range []string{
        `src="` + base + `/static/img/accessory/v2/active/humidity-sensor.png"`,
        `url.pathname.startsWith("/static/img/accessory/")`,
        `url.pathname = ingressBase + url.pathname;`,
        `new MutationObserver(function (mutations)`,
        `attributeFilter: ["src"]`,
        `document.addEventListener("error"`,
    } {
        if !strings.Contains(output, want) { t.Errorf("missing accessory image correction %q", want) }
    }
    if strings.Count(output, `src="` + base + `/static/img/accessory/`) != 1 {
        t.Error("static HTML image path must be prefixed once")
    }
}
