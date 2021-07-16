#!/usr/bin/env bash
hey -z 30s -c 50 \
  "http://autoscale-go.default.127.0.0.1.nip.io?sleep=100&prime=10000&bloat=5" \
  && kubectl get pods