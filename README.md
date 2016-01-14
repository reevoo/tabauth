# tabauth

tabauth is a thin wrapper service over Tableau Server's
[Trusted Authentication Endpoint](http://onlinehelp.tableau.com/current/server/en-us/help.htm#trusted_auth.htm%3FTocPath%3DAdministrator%2520Guide%7CTrusted%2520Authentication%7C_____0) 
that allows other servers, to authenticate using BasicAuth over https. tabauth removes any requirement for consuming applications to have static IP addresses.


## Usage

```
Usage of ./tabauth:
  -bind string
        the address to bind the tabauth server to (default "0.0.0.0:1443")
  -endpoint string
        the url for tableau server (default "http://localhost")
```

## Installation

These instuctions presume that you are running tabauth on the same server as Tableau Server. If you are running tabauth on another server, 
you will need to ensure that you have a static ip addess, and modify the instuctions for setting up Tableau Server appropriately.

This program is designed to be portable, so does not include any `Windows Service` functionality. We recommend running it with [NSSM](http://nssm.cc/).

Prebuilt binaries for Windows/amd64 are avalible on the [releases page](https://github.com/reevoo/tabauth/releases)

1. Create `C:\Program Files\tabauth\` (for example, put it wherever you like, but if its not here you may have to mess with permissions)
2. Copy `tabauth.exe` to `C:\Program Files\tabauth\`
3. Copy `cert.pem` and `key.pem` to `C:\Program Files\tabauth\`, you may be provided these by your CA or can [generate your own self-signed certificate](https://devcenter.heroku.com/articles/ssl-certificate-self)
4. Add accounts.json to `C:\Program Files\tabauth` [example](./accounts.json.example)
5. Setup the service using nssm:
`nssm install tabauth 'C:\Program Files\tabauth\tabauth.exe' '-endpoint=https://tableau.reevoo.com' '-bind=:1443'`

You may also want to adjust details like Logging etc:
`nssm edit tabauth`

Then start the service
`nssm start tabauth`

### Tableau Server Setup

In order for Tableau Server to "trust" tabauth, we need to configure it thus:

1. Get to the tabadmin command - `cd C:\Program Files\Tableau\Tableau Server\9.1\bin`
2. Stop tableau server - `tabadmin stop`
3. Set localhost as trusted - `tabadmin set wgserver.trusted_hosts "yourhost"`
4. Reload config files - `tabadmin config`
5. Restart tableau server `tabadmin start`

## Usage

You can request an access ticket for any user on your Tableau server:
```bash
$ curl --user user:password https://your-tableau-server/user/user-name/ticket
1gfuPluHbQbRv-VVNr44ecTH
```

You may restrict the tickets use by client IP address:
```bash
$ curl --user user:password https://your-tableau-server/user/user-name/ticket?client_ip=10.10.10.10
vgLDqQwHx_09iiUUDZwFPacZ
```
If want to do this you will need to configure Tableau Server to check the client IP on redeeming the ticket.

```
tabadmin set wgserver.extended_trusted_ip_checking true
tabadmin configure
tabadmin restart
```

If you are using a site other than the default one, you will need to specify a site id
```bash
$ curl --user user:password https://your-tableau-server/user/user-name/ticket?site_id=a4134fe9-d7ee-6783-88e9-a5eeb1f40476
vgLDqQwHx_09iiUUDZwFPacZ
```

## Development Requirements

You need go:

On OSX With Homebrew:
`brew install go`

Or follow these instructions for [other platforms](https://golang.org/doc/install)

## Testing

This project is tested

You can run the tests from the command line: 

```bash
$ go test
`

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

## Licence

This software is licenced under [The MIT License (MIT)](./LICENCE.md)
