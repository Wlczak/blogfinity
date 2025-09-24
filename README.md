# Blogfinity

![status](https://shokuin.vlastas.cc/api/badge/4/status?style=for-the-badge)
![uptime](https://shokuin.vlastas.cc/api/badge/4/uptime?style=for-the-badge)![phpstan](https://img.shields.io/github/actions/workflow/status/wlczak/blogfinity/.github%2Fworkflows%2Fgo-lint.yml?branch=main&style=for-the-badge&label=go-lint) ![build](https://img.shields.io/github/actions/workflow/status/wlczak/blogfinity/.github%2Fworkflows%2Fbuild.yml?branch=main&style=for-the-badge&label=build)

An infinitely generating pile of AI slop articles.
This project literaly doesn't serve any reasonable purpose, but like what do we need a purposefull project for anyway?

## Install by docker compose

```compose
services:
    blogfinity:
        image: wlczak/blogfinity:latest
            - ./db/:/app/db/
            - ./logger/logs/:/app/logger/logs/
        working_dir: /app
        restart: always
        extra_hosts:
            - "ollama-server:${OLLAMA_SERVER_IP}"
        ports:
            - "1111:8080"
        command: ./blogfinity
```

You will also need to set the OLLAMA_SERVER_IP environment variable to the IP of your Ollama server. For example like this:

```bash
# .env
OLLAMA_SERVER_IP=<ip>
BASE_DOMAIN=<domain>
```
