chm2docset
==========

[![Build Status](https://travis-ci.org/ngs/chm2docset.svg?branch=master)](https://travis-ci.org/ngs/chm2docset)
[![Coverage Status](https://coveralls.io/repos/github/ngs/chm2docset/badge.svg?branch=master)](https://coveralls.io/github/ngs/chm2docset?branch=master)


A tool for converting [Microsoft Compiled HTML Help (.chm)][chm] file to [Kapeli's Dash][dash] compatible Docset bundle

Usage
-----

```
usage: chm2docset [inputfile]
  -out string
        Output directory or file path (default "./")
  -platform string
        DocSet Platform Family (default "unknown")
```

How to use
----------

```sh
brew install chmlib
go install github.com/ngs/chm2docset
chm2docset -platform docset-platform -out /path/to/MyRef.docset /path/to/MyReference.chm
```

Author
------

[Atushi Nagase]

License
-------

MIT

[chm]: https://en.wikipedia.org/wiki/Microsoft_Compiled_HTML_Help
[dash]: https://kapeli.com/dash
[Atushi Nagase]: https://ngs.io/
