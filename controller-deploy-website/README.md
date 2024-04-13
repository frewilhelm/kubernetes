# Controller to deploy website

The `website-controller` is a Kubernetes controller that deploys websites from
Dockerimages as local websites.

This directory contains several files to build a Kubernetes controller (for
Minikube) that handle a custom resource `websites`. The controller creates
a deployment based on the custom resource that creates a Pod and a Service to
create a local website.

The static websites are stored in `websites/<name>/static`.

# How to use

First, one has to create the required Dockerimages. This can be achieved by
executing `buildDockerimages.sh`. First, the script configures to use Minikubes
local docker repository, so that the Pods can access them directly and no pull
is required. Then all Dockerimages are build by using the Dockerfiles in their
respective directories.

```sh
> ./buildDockerimages.sh
```

If the Dockerimages are present, the Deployment for the controller can be
applied.

```sh
> kubectl apply -f controller.yaml
```

This creates the deployment running the `website-controller` as well as the
`ClusterRole` definition and related resources.

After the `website-controller` is deployed, one can deploy websites by using
the custom resource `WebSite`.

```sh
> kubectl apply -f homepage-1.yaml
```

This will create the deployment consisting in a pod and service that provides
the local website. Check it out by executing `minikube service --all` to get
the respective IP.
