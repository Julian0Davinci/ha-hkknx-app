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
    js := `fetch("/api/status"); const x='/projects/x'; const y="https://external.tld";`
    out:=rewriteText(js,"application/javascript",base)
    if !strings.Contains(out,`fetch("`+base+`/api/status")`) || !strings.Contains(out,`x='`+base+`/projects/x'`) || !strings.Contains(out,`"https://external.tld"`) {t.Errorf("JS rewrite mismatch: %s",out)}
    css:=rewriteText(`background:url('/images/a.png')`,"text/css",base)
    if !strings.Contains(css,`url('`+base+`/images/a.png')`) {t.Errorf("CSS rewrite mismatch: %s",css)}
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
