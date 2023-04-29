
# upgrade to argocd v2.7-rc2
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/v2.7.0-rc2/manifests/install.yaml

# upgrade argocd cli as well
curl -sSL -o argocd-darwin-amd64 https://github.com/argoproj/argo-cd/releases/download/v2.7.0-rc2/argocd-darwin-amd64
sudo install -m 555 argocd-darwin-amd64 /usr/local/bin/argocd
rm argocd-darwin-amd64

# extension demo
gcloud builds submit --tag gcr.io/heewonk-bunker/gowiki

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

const myRequest = new Request("/extensions/httpbin/anything", {
  method: "GET",
  headers: myHeaders,
  mode: "cors",
  cache: "default",
});
const response = await fetch(myRequest);
console.log(await response.json())

const myRequest = new Request("/extensions/extdemo/view/test", {
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