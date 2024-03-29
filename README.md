# QrSync server

QR Sync is an application for syncronising across multiple computers.

It allows you to use a mobile phone to scan multiple other devices, and connect them for the purposes of sharing notes and files.

The QR Sync server is currently a go application that uses websockets to allow different clients to connect and send message to eachother.
It can be started by running 
```
go build && ./start.sh
```

It also uses typescriptify to allow message models defined a go structs, to be converted to Typescript defintions for use by the 
[web client](https://www.github.com/michaelclapham/qrsync-web). To generate these typescript definitions run
```
go build && ./qrsync-server -ts
```
This writes the defintions into models.ts which then needs to be copied over to qrsync-web manually.