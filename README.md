# ddnsd
DDNS daemon/client written in go!!

ONLY FOR NAMECHEAP AT THE MOMOENT BUT VERY EASY TO EXTEND.

Just add the protocol support in domain.useProtocol() and domain.checkError() methods.

##Install
install with

	go get github.com/nhooyr/ddnsd

It'll be installed to $GOPATH/bin, just make sure its in your path and you can launch it with

	ddnsd -c pathToConfigFIle

if no -c flag is given, the default location it looks for is /etc/ddnsd/config.json

example config.json file

<pre>
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
</pre>

Interval is in seconds.
