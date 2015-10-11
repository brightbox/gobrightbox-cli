# Brightbox Cloud Go CLI Client

`gobrightbox-cli` is a [Brightbox Cloud](https://www.brightbox.com)
[API](https://api.gb1.brightbox.com/1.0/) command line interface written in
[Go](http://golang.org/).

It uses the [`gobrightbox`](https://github.com/brightbox/gobrightbox) API Go
client library.

This tool is experimental and a work in progress. Not all the Brightbox services
are currently supported. For complete support of the Brightbox API, instead use
the [Ruby CLI](https://www.brightbox.com/docs/guides/cli/)

## Authentication

### User credentials

If you're logging in with your Brightbox user credentials, then use the `login` command:

    $ gobrightbox-cli login john@example.com
    Password for john@example.com:

This will obtain an OAuth refresh token and authentication token from the
Brightbox API and cache it locally, and add it as a "client" to the config.

The cached authentication token will work for 2 hours and the refresh token will
work for several more hours, after which point you'll get oauth2 authentication
errors and need to login again.

#### Multiple accounts

If you are a collaborator or owner of multiple accounts, a default account is
automatically selected for you and written to the config on the first run of the
`login` command.

You can change that default at any time by running the `login` command again
with the `--default-account` flag, and the config will be updated:

    $ gobrightbox-cli login --default-account acc-xxxxx john@example.com
    Password for john@example.com:

### API Client credentials

If you're planning to use the CLI from an automated system, you'll want to use
[API Client credentials](https://www.brightbox.com/docs/reference/api-clients/)
instead, which can be written to the config and will not expire (unless you
revoke the API Client).

You can add an API Client using the `config clients add` command:

    $ gobrightbox-cli config clients add --name=myaccount cli-aaaaa mysecret

And then use it like this:

    $ gobrightbox-cli --client=myaccount servers

## Compatibility with the Ruby CLI client

The Go CLI tool does not share a config file or token cache with the Ruby CLI,
they are kept separate.

The Go CLI user interface shares some similarities with the Ruby client but
differs in many ways and is not a drop-in replacement.

## Help

If you need help using this tool, drop an email to support at brightbox dot com.

## License

This code is released under an MIT License.

Copyright (c) 2015 Brightbox Systems Ltd.
