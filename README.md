# Hemono 荷物账本

## Prepare

```sh
# Create data dir
mkdir pb_data

# Setup PocketBase admin account 
cp .env.example .env
vim .env
```

## Develop

### Docker up
In the dev container, the backend uses Air to automatically monitor changes in Go code, while the frontend relies on Vite's HMR functionality.
```sh
docker compose -f docker-compose.dev.yml up -d
```

### Migrate database

If you change the database configuration, you will need to create a database migration. For details, refer to https://pocketbase.io/docs/go-migrations
```sh
go run . migrate collections
go run . migrate history-sync
```

## Deploy

### Docker up
```sh
docker compose -f docker-compose.yml up -d
```

### Reverse proxy with Caddy
Example Caddyfile

```Caddyfile
yourdomain.com {
    handle /api/* {
        reverse_proxy localhost:8090
    }

    handle /_/* {
        reverse_proxy localhost:8090
    }

    handle {
        reverse_proxy localhost:5173
    }

    encode zstd gzip
}
```
