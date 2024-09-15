SPIRE_SERVER=$(kubectl get pod -l app=spire-server -n spire -o jsonpath="{.items[0].metadata.name}")
echo "----------Bundle Show----------"
kubectl exec -it "$SPIRE_SERVER" -n spire -c spire-server -- /opt/spire/bin/spire-server bundle show -format spiffe
echo "----------Bundle List----------"
kubectl exec -it "$SPIRE_SERVER" -n spire -c spire-server -- /opt/spire/bin/spire-server bundle list -format spiffe
echo "----------Federation List----------"
kubectl exec -it "$SPIRE_SERVER" -n spire -c spire-server -- /opt/spire/bin/spire-server federation list