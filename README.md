# Quotes

A simple, self-hosted quotes website built with Go, HTMX, and Tailwind.

## About

My friend [Guney](https://github.com/gcg) started collecting quotes in 2006 during our IRC discussions on Freenode (#kanal and #turklug).

When he decided to shut down his quotes site, he backed up the entire database and gave it to me in 2015.

I’ve been storing that database ever since. It contains valuable memories from friends from nearly twenty years ago.

I thought it would be a shame to let it go to waste, so I decided to create a simple website to host the quotes.

Of course, I'm not sharing the entire database content, but I thought it would be nice to have a simple website to browse the quotes.

![](https://github.com/hionay/quotes/blob/main/images/quotes.jpg)

## Features

- Browse **latest**, **top**, and **random** quotes
- Add new quotes via a simple form
- Upvote or downvote existing quotes
- Responsive UI with Tailwind and dynamic interactions powered by HTMX

## Usage

Update configuration (port, DB connection) in `.env` (`.env.example` is provided as a template).

   ```shell
   cp .env.example .env
   ```

   > Note: The database should be created and seeded with the `quotes.sql` file.

Run the server:

   ```shell
   go build
   ./quotes
   ```
Visit `http://localhost:8080` in your browser.

## To launch with Docker Compose:

```shell
docker-compose up --build
```
