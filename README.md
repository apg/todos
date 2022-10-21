# TODOs: A sentinel on Tripods app ported to Go

A few hack weeks ago, I added a simple Web Framework to [sentinel](https://hashicorp.com/sentinel). No, it was never released. But! I now have a need
for a simple little app to demonstrate some other things. This is where
todos comes in. A simple, "real world"-ish app that is useful as a test bed
for other things. It's better than "hello world" and not much more to
maintain.

## Get this going.

You'll need to background some commands, or just use 3 terminals.

1. `make dev/vault` in terminal 1
2. `make dev/docker-db` in terminal 2
3. `./scripts/setup_local.sh`
4. Copy and paste the exports into the terminal from Step 3.
5. `../envbreach/envbreach ./todos` assuming you've got envbreach checked out and built
6. Browse to http://localhost:8080 and add some todos.
