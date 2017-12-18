# NOTHING TO SEE HERE

A pile of half-baked ideas.

## Development

The directory containing this file has to be on your file system (or
symlinked on your file system) as `$GOPATH/src/github.com/joeatwork/world-of-strategery`
or it won't build, and everything will go screwy.

Building the project requires go 1.8. To build, run

```
go build
```

To test, run

```
go test -cover ./game/
```

### Dependencies

Dependencies are managed with dep. To begin your development, run

```
dep ensure
```

This will pull a bunch of dependencies into your `vendor` directory.

