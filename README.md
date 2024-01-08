# powergrid

Like nginx, but for Discord bots.
![architecture diagram](https://i.sportshead.dev/xoyp3o3.jpg)

## setup
```bash
$ wget -O myvalues.yaml https://github.com/sportshead/powergrid/raw/master/powergrid/values.yaml
$ vim myvalues.yaml # set secrets, ingress
$ kubectl create namespace powergrid-development && kubectl config set-context --current --namespace=powergrid-development
$ helm repo add powergrid https://sportshead.github.io/powergrid
$ helm upgrade --install --wait -f myvalues.yaml powergrid powergrid/powergrid
```

## run an example bot
```bash
$ kubectl apply -f https://github.com/sportshead/powergrid/raw/master/examples/bun/bun.yaml
$ kubectl rollout restart deployment/powergrid # discover new CRD changes
```

## todo
- [x] CI
- [ ] more example bots
  - [ ] discord.js
  - [ ] discord.py
- [x] `shouldSendDeferred` option in `PowergridCommand` CRD
- [x] use typed kubernetes client in coordinator
- [x] watch for k8s `PowergridCommand` CRD changes
  - [x] reconcile changes with Discord
  - [x] support guild commands for development
- [ ] interaction routing - prefix style (`bun/*`) in a CRD/annotation?
