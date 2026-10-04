#!/bin/zsh
set -e

DIR="$( cd "$( dirname "$0" )" && pwd )"
CONTAINER_TAG=$(git rev-parse --short HEAD)
docker build --platform linux/amd64 -t gophprofile:$CONTAINER_TAG $DIR
docker save docker.io/library/gophprofile:$CONTAINER_TAG | ssh vla8islav@192.168.88.37 'sudo k3s ctr images import -'
sed -i '' "s|image: gophprofile:.*|image: gophprofile:${CONTAINER_TAG}|" "$DIR"/deploy/k8s/*.yaml

kubectl delete job gophprofile-migrate -n gophprofile --ignore-not-found
for f in "$DIR/deploy/k8s/00-namespace.yaml" "$DIR"/deploy/k8s/1[0-3]-*.yaml; do
  kubectl apply -f "$f"
done
kubectl apply -f "$DIR/deploy/k8s/15-migrate.yaml"
kubectl wait --for=condition=complete job/gophprofile-migrate -n gophprofile --timeout=120s
kubectl apply -f "$DIR/deploy/k8s/20-server.yaml"
kubectl apply -f "$DIR/deploy/k8s/21-worker.yaml"
kubectl apply -f "$DIR/deploy/k8s/30-ingress.yaml"
