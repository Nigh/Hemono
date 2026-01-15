# Hemono 荷物账本

## Prepare

```sh
# Create data dir
mkdir pb_data

# Setup PocketBase admin account 
cp .env.example .env
vim .env
```

## Start

### Develop

```sh
docker compose -f docker-compose.dev.yml up -d
```

### Production
```sh
docker compose -f docker-compose.yml up -d
```
