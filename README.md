gcloud builds submit --tag gcr.io/heewonk-bunker/gowiki

argocd app create argocd-extension-demo --repo https://github.com/saillinux/argocd-extension-demo.git --path helm --dest-server https://kubernetes.default.svc --dest-namespace argocd

argocd-extension-demo.argocd.svc.cluster.local
