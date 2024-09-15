SPIRE_SERVER=$(kubectl get pod -l app=spire-server -n spire -o jsonpath="{.items[0].metadata.name}")
kubectl exec -it "$SPIRE_SERVER" -n spire -c spire-server -- \
    /opt/spire/bin/spire-server entry show \