### Bender TODO
* Add some tests
* Switch to TOML for config file storage
* FInd a suitable datastore for factoids (yaml now, but probably not viable for long)
  - Also make importing possible from yaml
* Be able to import old bender's factoid db
  - If anyone wants to contribute an easy thing, make a parser that converts eggdrop factoid db to something structured, like json or yaml or whatever. Eggdrop's factoids are stored as:
```
    keyword => fact1 | fact2 | fact3
```
* Abstract config handling to avoid code duplucation
* Features:
  - ~Beatme~
  - weather
  - calculator
  - bar
  - factoids: reverse lookup
  - factoids: show additional info
  - factoids: list factoid keywords
  - irc: channel mode enforcing (low priority)
  - seen db, incl. if feasible away db
  - plugins? Overkill for now, probably
