# da

```
tail -F log | jq --unbuffered -cr '[ .state, .req.type ] | @tsv' | da
```
