chm2docset
==========

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
