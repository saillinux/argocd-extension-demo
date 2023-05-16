# Environment Setup

## GKE Cluster

### Workload Identity

We will need Workload Identity to allow our extension demo to access Google Cloud Compute API via its service account to control managed instance groups.

First of all, make sure your existing node pool is enabled with GKE Metadata.
```
GKE Cluster -> Node -> Node Pool -> Security -> Enable GKE Metadata Server
```

Create a service account for your extension application to use:

```
kubectl create serviceaccount sa-extdemo --namespace argocd

gcloud iam service-accounts create gsa-extdemo --project=heewonk-bunker

gcloud projects add-iam-policy-binding heewonk-bunker \
    --member "serviceAccount:gsa-extdemo@heewonk-bunker.iam.gserviceaccount.com" \
    --role "roles/compute.admin"

gcloud iam service-accounts add-iam-policy-binding gsa-extdemo@heewonk-bunker.iam.gserviceaccount.com \
    --role roles/iam.workloadIdentityUser \
    --member "serviceAccount:heewonk-bunker.svc.id.goog[argocd/sa-extdemo]"

kubectl annotate serviceaccount sa-extdemo \
    --namespace argocd \
    iam.gke.io/gcp-service-account=gsa-extdemo@heewonk-bunker.iam.gserviceaccount.com
```

When you create a pod or a deployment, add the following configuration to enable pods to use the service account

```
spec:
  serviceAccountName: sa-extdemo
  nodeSelector:
    iam.gke.io/gke-metadata-server-enabled: "true"
```

You can log in to the pod and check whether service account is associated with the pod
```
kubectl exec -it argocd-server-78b8784d4b-zgdgj -n argocd -- /bin/bash

curl -H"Metadata-Flavor: Google" http://metadata/computeMetadata/v1/instance/service-accounts/default/email    
```

The official guide for enabling the workload identity
- https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity
- https://spring-gcp.saturnism.me/deployment/kubernetes/workload-identity

## ArgoCD Setup

### Install ArgoCD along with ArgoCD Extension

https://github.com/argoproj-labs/argocd-extensions

```
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
# base Argo CD components
- https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/ha/install.yaml

components:
# extensions controller component
- https://github.com/argoproj-labs/argocd-extensions/manifests
```

use the following command to install the above kustomization

```
kustomize build . | kubectl apply -f - -n argocd
```

### Upgrade the existing to argocd v2.7.1
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/v2.7.1/manifests/install.yaml

### Upgrade argocd cli
curl -sSL -o argocd-darwin-amd64 https://github.com/argoproj/argo-cd/releases/download/v2.7.1/argocd-darwin-amd64
sudo install -m 555 argocd-darwin-amd64 /usr/local/bin/argocd
rm argocd-darwin-amd64

### Enable Extension Controller using the ConfigMap

Please visit the *GKE Secrets and ConfigMap* and select the ConfigMap *argocd-cm* then click Edit.

Copy paste the following to the configmap and please replace user1 to your username which will be used for login to ArgoCD with right permissions.

The following ConfigMap enables the newly deployed extension endpoint *extdemo* with proxy configuraiton and where the service is deployed so that ArgoCD Proxy can access the extension via its reverse proxy. The more detail is described in here https://argo-cd.readthedocs.io/en/latest/developer-guide/extensions/proxy-extensions/.


```
data:
  accounts.user1: apiKey,login
  extension.config: |-
    extensions:
    - name: extdemo
      backend:
        connectionTimeout: 3000000000
        keepAlive: 15000000000
        idleConnectionTimeout: 60000000000
        maxIdleConnections: 30
        services:
        - url: http://argocd-extension-demo.argocd.svc.cluster.local
          cluster:
            name: in-cluster
            server: https://kubernetes.default.svc
```

You will need to restart the argocd server

```
kubectl rollout restart deployment argocd-server -n argocd
```

### RBAC for extension access
https://argo-cd.readthedocs.io/en/latest/operator-manual/rbac/#the-extensions-resource

argocd admin settings rbac can admin get applications "default/argocd-extension-demo"  -n argocd
argocd admin settings rbac can admin invoke extensions extdemo -n argocd

create users by adding the following in argocd-cm ConfigMap (This was described in enabling ArgoCD extension already)

```
data:
  accounts.user1: apiKey,login  
```

update password of the user using the following command

```
argocd account update-password --account <user>
```

Update the *argocd-rbac-cm* ConfigMap to assign role to a user

```
data:
  policy.csv: |
    p, role:org-admin, applications, *, */*, allow
    p, role:org-admin, clusters, get, *, allow
    p, role:org-admin, repositories, get, *, allow
    p, role:org-admin, repositories, create, *, allow
    p, role:org-admin, repositories, update, *, allow
    p, role:org-admin, repositories, delete, *, allow
    p, role:org-admin, projects, get, *, allow
    p, role:org-admin, projects, create, *, allow
    p, role:org-admin, projects, update, *, allow
    p, role:org-admin, projects, delete, *, allow
    p, role:org-admin, logs, get, *, allow
    p, role:org-admin, exec, create, */*, allow
    p, role:org-admin, extensions, *, *, allow
    g, user1, role:org-admin
  policy.default: role:readonly
```

# Extension Demo

## Deploy Extension Demo

First checkout the extension demo repository and run the rest of command to build and upload image to your container registry (make sure you setup your own public container registry for this demo)

```
git clone https://github.com/saillinux/argocd-extension-demo.git

cd extdemo

# Build the container image using the Dockerfile.
gcloud builds submit --tag gcr.io/heewonk-bunker/extdemo

# Deploy the extension demo to argocd namespace so that ArgoCD Server can access the endpoint
argocd app create argocd-extension-demo --repo https://github.com/saillinux/argocd-extension-demo.git --path helm --dest-server https://kubernetes.default.svc --dest-namespace argocd
```

if you want to make sure the extension demo is installed correctly to test it locally.

```
kubectl port-forward svc/argocd-extension-demo -n argocd 8080:80
```

## Optional - Test from the ArgoCD UI

You can use the console in Chrome Developer Tools using the following javascript snippet. You can obtain the cookie token for accessing the backend from any request made from the UI.

```
const myHeaders = new Headers();
myHeaders.append("Cookie", "")
myHeaders.append("Argocd-Application-Name", "argocd:argocd-extension-demo")
myHeaders.append("Argocd-Project-Name", "default")

const myRequest = new Request("/extensions/extdemo/compute/instancegroup/heewonk-bunker/us-west1/heewonk-bunker-argocd-cc-demo-instance-group", {
  method: "GET",
  headers: myHeaders,
  mode: "cors",
  cache: "default",
});
const response = await fetch(myRequest);
console.log(await response.json())
```

## Install extension UI example

Use the following ArgoCDExtension yaml to have the extension controller to download the new UI. The extension UI build is uploaded to the github release and the controller will download and extract the files to /tmp/extension to enable the UI for *ManagedInstanceGroup* Resource.

https://github.com/saillinux/argocd-extension-demo/blob/main/argocd_extension.yaml
```

apiVersion: argoproj.io/v1alpha1
kind: ArgoCDExtension
metadata:
  name: argocd-extension-example
  labels:
    tab: "Demo"
    icon: "fa-th"
  finalizers:
    - extensions-finalizer.argocd.argoproj.io
spec:
  sources:
    - web:
        url: https://github.com/saillinux/argocd-extension-demo/releases/download/v0.1.2/extension34.tar

```

The pre-built extension file is located in here https://github.com/saillinux/argocd-extension-demo/releases/tag/v0.1.2. You can build the ui by following the local development section.


Now apply the change then the controller will install them in the right directory and access the UI once more e.g. instancegroupmamaner resource More Tab.
```
kubectl apply -f argocd_extension.yaml -n argocd
```

# Local Development

```
# Use ADC for Google Cloud credential for your project
gcloud auth application-default login

# Obtain the GKE cluster credential
gcloud container clusters get-credentials front01-us-central1-c --zone us-central1-c

# Checkout the extension demo repository
git clone https://github.com/saillinux/argocd-extension-demo.git

# For backend development
cd extdemo
go mod download
go build -o extdemo .

# For UI development
cd ui
yarn build

# copy the extension.tar from dist direcotry to the bucket or github release and use the following extension yaml to have the extension controller to download the new UI. Make sure you have pointed to the right url where the UI extension is uploaded

vi ../argocd_extension.yaml

kubectl apply -f ../argocd_extension.yaml
```

## Notes

The Group/Kind value in webpack.config.js is used to associate which resource this UI component will be displayed as a More Tab in the UI.

```
const groupKind = 'compute.cnrm.cloud.google.com/ComputeInstanceGroupManager';
```

# TBD - Enable Custom Application View

Make sure the Group/Kind is *argoproj.io/Application* in webpack.config.js

((window) => {
  window.extensionsAPI.registerAppViewExtension(
    window.extensions.resources["argoproj.io/Application"].component,
    window.extensions.resources["argoproj.io/Application"].title,
    window.extensions.resources["argoproj.io/Application"].icon,
  );
})(window);

# References
https://pkg.go.dev/google.golang.org/api/compute/v1

https://cloud.google.com/compute/docs/reference/rest/v1/regionInstanceGroupManagers/get
https://cloud.google.com/compute/docs/reference/rest/v1/regionInstanceGroups/listInstances

https://cloud.google.com/compute/docs/instance-groups/rolling-out-updates-to-managed-instance-groups
https://cloud.google.com/compute/docs/reference/rest/v1/regionInstanceGroupManagers/patch#iam-permissions

The webpack config contains the group kind
- https://github.com/argoproj-labs/argocd-extension-metrics/blob/main/extensions/resource-metrics/resource-metrics-extention/ui/webpack.config.js
- https://github.com/argoproj-labs/argocd-example-extension/blob/master/ui/webpack.config.js
