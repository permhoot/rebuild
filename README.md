# rebuild

Fun project to evaluate the complexity of the idea to have a Knative application that serves as a GitHub webhook target to trigger Shipwright build in order to update the Knative application that is using the respective container image. Simply put, what is the bare minimum glue code required to update an existing Knative application when there is new code.

## Pre-requisites

Kubernetes cluster with [Shipwright](https://github.com/shipwright-io/build/) as the container image building framework and [Knative Serving](https://knative.dev/docs/serving/) as the application serving framework, for example like the [IBM Cloud® Code Engine](https://www.ibm.com/products/code-engine) service.

## Setup

This is a test project, therefore there is no convenience script or setup routine. For example, with Code Engine, make sure you log into your account, select the project and then use `ko` to build and the IBM Cloud CLI with the Code Engine plugin to roll out the tool.

```sh
KO_DOCKER_REPO=some.container.registry/namespace/image ko build --bare --sbom=none |
  xargs --no-run-if-empty -I{} ibmcloud ce app update --name rebuild --image {}
```

## How does it work

Place a Build in a namespace to define the source, intended build strategy, and the output image. The Git source defined in the build will be the identifying reference for the tool to figure out, which build needs to be started. It will also work with a BuildRun that has an embedded BuildSpec, also known as standalone BuildRuns.

Create a Knatitve service that is using the same image reference as the one that is defined in the Build specification output. This will be the identifying reference for the tool to figure out, which Knative service needs a new revision in order to roll out the new image version to be served.

That's it in the cluster.

Next, you need to configure the URL of the tool in the GitHub repository _Settings_ under _Webhooks_: Click _Add Webhook_ and put the tool URL as the _Payload URL_. Select _Branch or tag creation_ for example if this is only for new tags or releases. Save it by clicking _Add Webhook_.

Whenever GitHub fires the Webhook payload URL, the tool will take the event and start to look for a Build that references the GitHub repository URL from the event payload. When found, it will trigger a new BuildRun based on that Build and wait for it to complete. After it is completed successfully, it will nudge the Knative Service to pick up the new container image to create a new revision, which will roll out the new version.
