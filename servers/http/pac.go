package http

import (
	"io"
	"net/http"
	"text/template"
)

const TEMPLATE string = `
var proxy = new Array();

{{ range $proto, $addr := .Proxies }}
proxy["{{ $proto }}"] = "{{ $addr }}";
{{ end }}

function FindProxyForURL(url, host)
{
	if ( shExpMatch(host, "({{ range $domain := .Domains }}*.{{ $domain }}|{{ $domain }}|{{ end }}doxy)") ) {
		if ( "socks" in proxy ) {
			return "SOCKS5 " + proxy["socks"];
		} else if ( url.substring(0, 5) == "http:" ) {
			//return "PROXY " + proxy["http"] + "; DIRECT";
			return "PROXY " + proxy["http"];
		} else if ( url.substring(0, 6) == "https:" ) {
			return "PROXY " + proxy["https"];
		}
	}

	// Last resort, go direct
	return "DIRECT";
}
`

type PACTemplateContext struct {
	Proxies map[string]string
	Domains []string
}

func (s *HTTPProxy) generatePAC(buf io.Writer) error {
	tmpl, err := template.New("pac").Parse(TEMPLATE)
	orPanic(err)

	ctx := PACTemplateContext{
		Domains: make([]string, 0),
		Proxies: make(map[string]string, 0),
	}

	// TODO This shouldn't be static like this.
	ctx.Proxies["http"] = s.config.HttpAddr
	ctx.Proxies["https"] = s.config.HttpsAddr

	// SOCKS5(h) must be used in the case of dns not going to doxy.
	ctx.Proxies["socks"] = s.config.SocksAddr

	//ctx.Domains = append(ctx.Domains, "doxy.docker")
	ctx.Domains = append(ctx.Domains, "docker")

	err = tmpl.Execute(buf, ctx)
	return err
}

func (s *HTTPProxy) handlePAC(w http.ResponseWriter, r *http.Request) {
	logger.Debugf("Generating PAC file")
	s.generatePAC(w)
}
