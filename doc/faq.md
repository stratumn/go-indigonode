# Frequently Asked Questions

## When did Alice start?

October 12th 2017, walking back home from work (@rcaetano and @stratumn).

## Why ports 8903-8904?

Because `890` is easy to type on a keyboard, and ports `8903` to `8907` are
neither registered on [iana.org](https://www.iana.org) nor detected by
[nmap](https://svn.nmap.org/nmap/nmap-services), giving us room for more ports
if we need them without risking of conflicting with other applications.

## Why a CLI instead of terminal commands?

Because the CLI uses reflection to automatically create commands from the API.
It is possible that the API has functions the `alice` binary doesn't know
about when it's compiled, so it cannot have commands for them. You can use
`alice -c command` though.
