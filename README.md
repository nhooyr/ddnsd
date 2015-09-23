# goDDNS
DDNS client written in go!!

working on adding more support for protocols, only supports namecheap atm.

##Install
install with

	go get github.com/aubble/goDDNS

It'll be installed to $GOPATH/bin, just make sure its in your path and you can launch it with

	goDDNS -c pathToConfigFIle

if no -c flag is given, the default location it looks for is /etc/goDDNS/config.json

example config.json file

{
  "List": [
    {
      "Host": "@",
      "Domain": "aubble.com",
      "Password": "pass",
      "Protocol": "namecheap"
    },
    {
      "Host": "d",
      "Domain": "aubble.com",
      "Password": "pass",
      "Protocol": "namecheap"
    }
  ],
  "Interval": 180
}

Interval is in seconds.

Very easy to extend, just add the protocol support in domain.useProtocol() and domain.checkError() methods.
