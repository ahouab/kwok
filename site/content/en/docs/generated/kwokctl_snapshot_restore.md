## kwokctl snapshot restore

Restore the snapshot of the cluster

```
kwokctl snapshot restore [flags]
```

### Options

```
  -h, --help          help for restore
      --path string   Path to the snapshot
```

### Options inherited from parent commands

```
  -c, --config strings   config path (default [~/.kwok/kwok.yaml])
      --dry-run          Print the command that would be executed, but do not execute it
      --name string      cluster name (default "kwok")
  -v, --v log-level      number for the log level verbosity (DEBUG, INFO, WARN, ERROR) or (-4, 0, 4, 8) (default INFO)
```

### SEE ALSO

* [kwokctl snapshot](kwokctl_snapshot.md)	 - Snapshot [save, restore, record, replay, export] one of cluster

