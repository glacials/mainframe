# mainframe

Mainframe is a monolith containing some convenience home-related code that I
run and pipe data into and out of. It's a replacement for a series of brittle
shell scripts.

If you can find some use out of it, great!

## Running persistently

To run mainframe how it's meant to be run, e.g. on a Raspberry Pi in a closet, first
install dependencies:

```sh
sudo apt-get install moreutils
```

Then set up a user cron like so:

```cron
0 0 * * * chronic bash /path/to/mainframe/supervisor.sh
```
