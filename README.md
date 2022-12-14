# mainframe

Mainframe is a monolith containing some personal automations and convenience
code that I run 24/7 and pipe data into and out of. It's a replacement for a
series of brittle shell scripts and cron jobs.

If you can find some use out of it, great!

## Running as a daemon

To run mainframe how it's meant to be run, e.g. on an old machine or Raspberry
Pi in a closet, first install dependencies:

```sh
sudo apt-get install moreutils
```

Then set up a user cron like so:

```cron
0 0 * * * chronic bash /path/to/mainframe/supervisor.sh
```

The supervisor script will handle the rest, including auto-updating.
