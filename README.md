## CALLER

This (quite dumb) caller calls jsonrpc methods with given RPS.

Usage:

    caller rps num-connections url header-file request-file id-list-file

`header-file` must be a text file with HTTP headers, one header per line
For example:

    Authorization: basic, 123412341234

`request-file` is a (probably not syntactically correct) json file

    {
        "0.1": {lol": "#VAL#"},
        "0.2": {"kek": "#FOO#"},
        "0.666": {"id": "#ID#", "cheburek": #BAR#}
    }

The keys here are probabilities of each request to appear.
(the probabilities may not add up to 1.0)

`id-list-file` is a (now valid) json with lists of identifiers:

    {
        "VAL": ["1", "2", "3"],
        "FOO": ["foo", "bar", "baz"],
        "BAR": ["666", "999"]
    }

Note that the keys here do match with `#XXX#` templates in the `request-file`.

When a random request is to be sent, it is scanned for template sequences and
each such sequence is substituted with a random value from the corresponding list.
A special pattern `#ID#` is replaced by a random string.

So in our example the first request `{"lol": "#VAL#"}` will take the form
`{"lol": "1"}`, `{"lol": "2"}` or `{"lol": "1"}` (at random).

The substitution is performed on string level and does NOT pay attention to json syntax
(therefore the json in `request-file` may not be a valid json, like in the example above,
but it will hopefully become a correct one after the substitution).

