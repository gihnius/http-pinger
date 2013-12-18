# http pinger

A simple tool to check website status and notify via email.

## build executable
`go build http-pinger.go`

## run from console
`./http-pinger `

OR:
```
go run http-pinger.go
```

## config.json

```
{
    "lag": 30,
    "interval": 60,
    "urls_file": "urls.txt",
    "smtp_username": "",
    "smtp_password": "",
    "smtp_host": "localhost",
    "smtp_port": "25",
    "email_subject": "http pinger",
    "from_email": "admin@server.com",
    "to_emails": ["admingroup@server.com"]
}
```
Alerts by email when a url's response time longer than **lag** in seconds.


## urls file

```
http://google.com/
http://baidu.com/
https://non-exist.xyz/
```
lines without http:// or https:// beginning would be ignored.

## TODO

* monitor other services: smtp, ftp, ping, pop3, custom port...
* check spectial expected http status code for a url
* add notify by twitter, SMS ?


## License

The MIT License (MIT)
