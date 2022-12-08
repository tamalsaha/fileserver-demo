# fileserver-demo

```
curl -X POST -F file=@/path/to/file.ext http://localhost:8100
```

## kubectl-curl

- https://github.com/segmentio/kubectl-curl

```
./kubectl-curl_v0.1.5_darwin_arm64 -k https://scanner-0:8443/files/ -n kubeops

./kubectl-curl_v0.1.5_darwin_arm64 -k -X POST -F file=@/opt/homebrew/bin/kubectl  https://scanner-0:8443/files/a/b -n kubeops
```
