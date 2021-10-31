#!/usr/bin/env bash
hey -z 120s -c 10 \
  "http://autoscale-go.default.127.0.0.1.nip.io?sleep=100&prime=10000&bloat=5" \
  && kubectl get pods