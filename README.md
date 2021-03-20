# da

```
$ tail -F some.log | jq --unbuffered -cr '[ .country, .place, .device.os ] | @tsv' | da
```

<img src="./demo.gif" width="100%">

## Options

```
  --delimiter string
         (default "\t")
  --ignore-empty
        Do not show blank value count in a bar chart. (default true)
  --numeric-sort
        Sort the horizontal axis of a bar chart by numeric. (default true)
  --redraw-every int
        Redraws the screen cleanly every specified number of seconds. (default 5)
```

## Hot Keys

* `q` - quit app
* `r` - reset ui

## See Also

* https://github.com/keithknott26/datadash
* https://github.com/atsaki/termeter