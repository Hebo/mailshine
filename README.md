# Mailshine

Like [Mailbrew](https://mailbrew.com), but rss compatible

Get a daily digest of your favorite subreddits and Twitter posters that you can read in your RSS reader.

**Warning:** Very early WIP, not production ready!


## Development


Installing

```
go get github.com/hebo/mailshine
```

Run server with auto-reload (requires ripgrep and entr)

```
rg --files -g '*.{go,tpl}' | entr -r go run ./cmd/mailshine
```

Trigger generation

```
docker exec -it 7e96586d6af6 /app/mailshine -generate
```

Switch entrypoint
```
docker run --rm --entrypoint /bin/bash -it mailshine
```
