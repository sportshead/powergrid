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
$ kubectl apply -f https://github.com/sportshead/powergrid/raw/master/examples/bun/bunbot.yaml
$ kubectl rollout restart deployment/powergrid # discover new CRD changes
```

## todo
- [ ] CI
- [ ] more example bots
  - [ ] discord.js
  - [ ] discord.py
- [ ] `shouldSendDeferred` option in `PowergridCommand` CRD
- [ ] use typed kubernetes client in coordinator
- [ ] watch for k8s `PowergridCommand` CRD changes
  - [ ] reconcile changes with Discord
  - [ ] support guild commands for development
- [ ] interaction routing - prefix style (`bun/*`) in a CRD/annotation?
