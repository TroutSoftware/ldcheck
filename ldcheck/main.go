package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"strings"
	"text/template"

	"crypto/tls"

	"github.com/BurntSushi/toml"
	"github.com/go-ldap/ldap/v3"
	"golang.org/x/term"
)

func TestLoginWithLDAP(url, query, name, pass string) {
	tpl, err := template.New("ldap").Parse(query)
	if err != nil {
		log.Fatal(err, "invalid bind pattern, no access will ever match")
	}

	ctn, err := DialURL(url)
	if err != nil {
		log.Fatal(err, "cannot contact LDAP server")
	}

	var userdn strings.Builder
	if err := tpl.Execute(&userdn, struct{ UserName string }{ldap.EscapeFilter(name)}); err != nil {
		log.Fatal(err, "invalid LDAP query pattern")
		return
	}
	fmt.Println("checking user DN:", userdn.String())

	if err := ctn.Bind(userdn.String(), pass); err != nil {
		log.Fatal(err, "connection denied")
		return
	}

	userq := ldap.NewSearchRequest(
		userdn.String(),
		ldap.ScopeBaseObject, ldap.NeverDerefAliases, 0, 0, false,
		"(&)",
		[]string{"displayName", "mail"}, // A list attributes to retrieve
		nil,
	)

	sr, err := ctn.Search(userq)
	if err != nil {
		log.Fatal(err, "invalid user record in LDAP: contact your system administrator")
		return
	}

	for _, entry := range sr.Entries {
		entry.PrettyPrint(2)
	}
}

func readPassword() (s []byte, err error) {
	if term.IsTerminal(int(os.Stdin.Fd())) {
		fmt.Fprint(os.Stdin, "password ")
		return term.ReadPassword(int(os.Stdin.Fd()))
	}
	panic("not implemented")
}

func main() {
	var (
		conf = flag.String("file", "ldap.local.toml", "Configuration file to check")
		name = flag.String("name", "johndoe", "User Name")
		pass = flag.String("pass", "correcthorsebatterystaple", "Password")
	)
	flag.Parse()

	var Config struct {
		LDAP struct {
			ServerURL       string `toml:"server_url"`
			BindUserPattern string `toml:"bind_pattern"`
		}
	}
	_, err := toml.DecodeFile(*conf, &Config)
	if err != nil {
		log.Fatal(err)
	}

	TestLoginWithLDAP(Config.LDAP.ServerURL, Config.LDAP.BindUserPattern, *name, *pass)
}

func DialURL(addr string) (*ldap.Conn, error) {
	lurl, err := url.Parse(addr)
	if err != nil {
		return nil, ldap.NewError(ldap.ErrorNetwork, err)
	}

	host, port, err := net.SplitHostPort(lurl.Host)
	if err != nil {
		// we asume that error is due to missing port
		host = lurl.Host
		port = ""
	}

	switch lurl.Scheme {
	case "ldapi":
		if lurl.Path == "" || lurl.Path == "/" {
			lurl.Path = "/var/run/slapd/ldapi"
		}
		return ldap.Dial("unix", lurl.Path)
	case "ldap":
		if port == "" {
			port = ldap.DefaultLdapPort
		}
		return ldap.Dial("tcp", net.JoinHostPort(host, port))
	case "ldaps":
		if port == "" {
			port = ldap.DefaultLdapsPort
		}
		tlsConf := &tls.Config{
			ServerName: host,
			MinVersion: tls.VersionTLS10,
		}
		return ldap.DialTLS("tcp", net.JoinHostPort(host, port), tlsConf)
	}

	return nil, ldap.NewError(ldap.ErrorNetwork, fmt.Errorf("Unknown scheme '%s'", lurl.Scheme))
}
