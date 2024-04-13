# go-client-watcher

A little project to learn how to interact with Kubernetes using go-client.

The program watches for any Pod-related event and prints the event, Pod-Name,
and Namespace to the command-line.

Minikube is used as Kubernetes. If one uses the "normal" Kubernetes the
program has to be reconfigured by switching the configuration load (see
`main.go` commented lines).

## How to build

```sh
> go build -o pod-watcher
```

## How to run

```sh
> ./pod-watcher
```

## How to see output

```sh
> kubectl apply -f pod.yaml
```

Output:
```
...
Pod created (default): my-pod
Pod modified (default): my-pod
Pod modified (default): my-pod
Pod modified (default): my-pod
```

Press `CTRL+C` to terminate the `pod-watcher` or send `SIGINT`/`SIGTERM` to the
process.
