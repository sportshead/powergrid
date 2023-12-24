# powergrid

Like nginx, but for Discord bots.
![architecture diagram](https://i.sportshead.dev/xoyp3o3.jpg)

## setup
```bash
$ cp ./powergrid/values.yaml myvalues.yaml
$ vim myvalues.yaml # set secrets, ingress
$ kubectl create namespace powergrid-development && kubectl config set-context --current --namespace=powergrid-development
$ docker build -t ghcr.io/sportshead/powergrid-coordinator:0.1.0 ./coordinator/
$ helm upgrade --install -f myvalues.yaml powergrid ./powergrid
```

## run an example bot
```bash
$ docker build -t ghcr.io/sportshead/powergrid-examples-bunbot:1.0.0 ./examples/bun
$ kubectl apply -f ./examples/bun/bunbot.yaml
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
