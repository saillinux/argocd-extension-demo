
# upgrade to argocd v2.7-rc2
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/v2.7.0-rc2/manifests/install.yaml

# upgrade argocd cli as well
curl -sSL -o argocd-darwin-amd64 https://github.com/argoproj/argo-cd/releases/download/v2.7.0-rc2/argocd-darwin-amd64
sudo install -m 555 argocd-darwin-amd64 /usr/local/bin/argocd
rm argocd-darwin-amd64

# extension demo
gcloud builds submit --tag gcr.io/heewonk-bunker/extdemo

kubectl port-forward svc/argocd-extension-demo -n argocd 8080:80

argocd app create argocd-extension-demo --repo https://github.com/saillinux/argocd-extension-demo.git --path helm --dest-server https://kubernetes.default.svc --dest-namespace argocd

argocd-extension-demo.argocd.svc.cluster.local

kubectl rollout restart deployment argocd-server -n argocd

```
data:
  accounts.heewonk: apiKey,login
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
    - name: httpbin
      backend:
        connectionTimeout: 2000000000
        keepAlive: 15000000000
        idleConnectionTimeout: 60000000000
        maxIdleConnections: 30
        services:
        - url: http://httpbin.org
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


# RBAC for extension access
https://argo-cd.readthedocs.io/en/latest/operator-manual/rbac/#the-extensions-resource

argocd admin settings rbac can admin get applications "default/argocd-extension-demo"  -n argocd
argocd admin settings rbac can admin invoke extensions extdemo -n argocd

argocd-rbac-cm

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
    g, heewonk, role:org-admin
  policy.default: role:readonly
```

# Workload Identity
https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity

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

spec:
  serviceAccountName: sa-extdemo
  nodeSelector:
    iam.gke.io/gke-metadata-server-enabled: "true"        

# install argocd-extentions
kustomization.yaml

```
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
# base Argo CD components
- https://raw.githubusercontent.com/argoproj/argo-cd/v2.7.0-rc2/manifests/install.yaml

components:
# extensions controller component
- https://github.com/argoproj-labs/argocd-extensions/manifests
```

kustomize build . | kubectl apply -f - -n argocd

# install extension UI example
https://github.com/argoproj-labs/argocd-example-extension

webpack config contains the group kind
- https://github.com/argoproj-labs/argocd-extension-metrics/blob/main/extensions/resource-metrics/resource-metrics-extention/ui/webpack.config.js
- https://github.com/argoproj-labs/argocd-example-extension/blob/master/ui/webpack.config.js

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
        url: https://github.com/saillinux/argocd-extension-demo/releases/download/v0.1.0/extension.tar

# App View

((window) => {
  window.extensionsAPI.registerAppViewExtension(
    window.extensions.resources["argoproj.io/Application"].component,
    window.extensions.resources["argoproj.io/Application"].title,
    window.extensions.resources["argoproj.io/Application"].icon,
  );
})(window);

# scratch

without using ArgoCD extensions proxy
https://github.com/argoproj-labs/argocd-extension-metrics#enable-the-argo-ui-to-access-the-argocd-metrics-server

access the container
kubectl exec -it argocd-server-78b8784d4b-zgdgj -n argocd -- /bin/bash

# Google Cloud Go SDK for Managed Instance Group
https://pkg.go.dev/google.golang.org/genproto/googleapis/cloud/compute/v1
https://raw.githubusercontent.com/google/google-api-go-client/master/compute/v1/compute-gen.go

https://cloud.google.com/compute/docs/reference/rest/v1/regionInstanceGroupManagers/get
https://cloud.google.com/compute/docs/reference/rest/v1/regionInstanceGroups/listInstances

https://cloud.google.com/compute/docs/instance-groups/rolling-out-updates-to-managed-instance-groups
https://cloud.google.com/compute/docs/reference/rest/v1/regionInstanceGroupManagers/patch#iam-permissions