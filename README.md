GoWiki
=====

This is a simple wiki that I am working on to learn some basic concepts of Go.
It is quite simple and nothing fancy yet. It stores page in markdown inside the
`data` directory. It supports templates and also nested templates.

To compile this program run:

```bash
$ go build wiki.go
```

Then you can run the program using:

```bash
$ ./wiki
```

This starts an HTTP server listening at [localhost:8000]
