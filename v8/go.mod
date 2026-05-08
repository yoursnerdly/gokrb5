module github.com/yoursnerdly/gokrb5/v8

replace github.com/jcmturner/gokrb5/v8 => .

go 1.25.0

require (
	github.com/gorilla/sessions v1.4.0
	github.com/hashicorp/go-uuid v1.0.3
	github.com/jcmturner/aescts/v2 v2.0.0
	github.com/jcmturner/dnsutils/v2 v2.0.0
	github.com/jcmturner/gofork v1.7.6
	github.com/jcmturner/goidentity/v6 v6.0.1
	github.com/jcmturner/gokrb5/v8 v8.4.4
	github.com/jcmturner/rpc/v2 v2.0.3
	github.com/stretchr/testify v1.11.1
	golang.org/x/crypto v0.50.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/gorilla/securecookie v1.1.2 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/net v0.53.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
