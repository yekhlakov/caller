## CALLER

This (quite dumb) caller calls jsonrpc methods with given RPS.

Usage:

    caller [config-file-name]

If the config file name is not provided, `config.json` is assumed.

### config file

The config file is a json of the following structure:

    {
        "rps": 666,
        "numConnections": 100,
        "url": "https://host.name/path",
        "headers": {
            "Authorization": "basic, 123412341234"
        },
        "requestTemplates": [
            {"probability": 0.1, "template": {lol": "##VAL##"}},
            {"probability": 0.2, "template": {"kek": "#FOO#"}},
            {"probability": 0.666, "template": {"id": "#ID#", "cheburek": "#BAR#"}}
        ],
        "idLists": {
            "VAL": ["1", "2", "3"],
            "FOO": ["foo", "bar", "baz"],
            "BAR": ["666", "999"]
        }
    }

#### `headers` section

This section contains headers to be passed along with every request.
The header `Content-Type: application/json` is added by the caller.

#### `requestTemplates` section

This section contains the templates for the requests to be sent to the server.
The requests are picked at random with the specified probabilities
(the probabilities of all templates may not add up to 1.0).

#### `idLists` section

Note that the keys of this section do match with `#XXX#` fragments in the request templates.

When a random request is to be sent, it is scanned for template sequences and
each such sequence is substituted with a random value from the corresponding list.
(A special pattern `#ID#` is replaced by a random string).

The pattern `#XXX#` is replaced as is. The pattern `"##XXX##"` is replaced including
the quotes (which may be useful for inserting integers and other values outside json strings).

So in our example the first request `{"lol": "##VAL##"}` will take the form
`{"lol": 1}`, `{"lol": 2}` or `{"lol": 3}` (at random).
