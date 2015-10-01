# tabauth

Tabauth is a thin proxy over Tableau Server's [Trusted Authentication Endpoint](http://onlinehelp.tableau.com/current/server/en-us/help.htm#trusted_auth.htm%3FTocPath%3DAdministrator%2520Guide%7CTrusted%2520Authentication%7C_____0)
It allows other servers, with access to authenticate using BasicAuth over https. This removes any requirement for our servers to have static IP addresses.

## Usage

You can request an access token for any user on your Tableau server:
```bash
$ curl --user user:password https://your-tableau-server/user/user-name/token
1gfuPluHbQbRv-VVNr44ecTH
```

You may restrict the tokens use by client IP address:
```bash
$ curl --user user:password https://your-tableau-server/user/user-name/token?client_ip=10.10.10.10
vgLDqQwHx_09iiUUDZwFPacZ
```

If you are using a site other than the default one, you will need to specify a site id
```bash
$ curl --user user:password https://your-tableau-server/user/user-name/token?site_id=a4134fe9-d7ee-6783-88e9-a5eeb1f40476
vgLDqQwHx_09iiUUDZwFPacZ
```

## Development

You need go:

`brew install go`

## Testing

This project is tested with [GoConvey](http://goconvey.co/).

You can run the tests -

From the command line:
`go test`

In the browser:
```bash
go get github.com/smartystreets/goconvey
$GOPATH/bin/goconvey
```
Then visit [localhost:8080](http://localhost:8080)

## Building

To build the `.exe` for windows:

Make sure you have go 1.5+ installed:

```bash
$ go version
go version go1.5 darwin/amd64
```

Then you can cross compile for windows, by setting GOOS and GOARCH appropriately 

[see possible values...](https://github.com/golang/go/blob/master/src/go/build/syslist.go):

```bash
$ env GOOS=windows GOARCH=amd64 go build tabauth.go
```

## Installation

This program is designed to be portable, so does not include any `Windows Service` functionality. We recommend running it with [NSSM](http://nssm.cc/).

1. Create `C:\Program Files\tabauth\` (for example, put it wherever you like, but if its not here you may have to mess with permissions)
2. Copy `tabauth.exe` to `C:\Program Files\tabauth\`
3. Copy `cert.pem` and `key.pem` to `C:\Program Files\tabauth\`
4. Add accounts.json to `C:\Program Files\tabauth`
```json
{
  "user": "password",
  "otheruser": "sekret"
}
```
4. Setup the service using nssm:
```
nssm install tabauth C:\Program Files\tabauth\tabauth.exe
nssm start tabauth
```
The defaults should be fine, but you may want to adjust details like Logging etc:
`nssm edit tabauth`

### Tableau Server Setup

In order for Tableau Server to "trust" tabauth, we need to configure it thus:

1. Get to the tabadmin command - `cd C:\Program Files\Tableau\Tableau Server\9.1\bin`
2. Stop tableau server - `tabadmin stop`
3. Set localhost as trusted - `tabadmin set wgserver.trusted_hosts "127.0.0.1"`
4. Reload config files - `tabadmin config`
5. Restart tableau server `tabadmin start`
