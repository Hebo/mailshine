# Mailshine

Like [Mailbrew](https://mailbrew.com), but RSS compatible.

Get a daily digest of your favorite subreddits and Twitter posters that you can peruse in your RSS reader.

**Warning:** Very early WIP, not production ready!

![Screenshot](https://github.com/Hebo/mailshine/raw/main/resources/preview_screenshot.png)


## Usage

Docs to come eventually...

1. Set `.env` and `config.toml`
2. Deploy Dockerfile

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
