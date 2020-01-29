[![Docker Pulls](https://img.shields.io/docker/pulls/kpurdon/iapap.svg)](https://hub.docker.com/r/kpurdon/iapap)
![GitHub tag (latest SemVer)](https://img.shields.io/github/v/tag/kpurdon/iapap?sort=semver)

iapap
-----

[Identity Aware Proxy](https://cloud.google.com/iap/) Auth Proxy is a simple reverse proxy that implements the Google Cloud IAP [authentication requirements](https://cloud.google.com/iap/docs/signed-headers-howto). It's intended to be used as a kubernetes sidecar proxying to a target localhost service in the case where you cannot add authentication to the target service.

## Configuration

IAPAP is configured by the following environment variables:

- IAPAP_LOG_LEVEL: The log level DEBUG,INFO,WARN,ERROR (default: INFO)
- IAPAP_PORT: The port the IAPAP service will listen on (default: 8000)
- IAPAP_TARGET: The target service IAPAP will proxy too (default: http://localhost:8001)
- IAPAP_AUDIENCE: The "Signed Header JWT Audience" from Cloud IAP (required, no default)
- IAPAP_ENDPOINT_WHITELIST: An optional comma separated list of endpoints to proxy unauthenticated. Useful for allowing healthcheck endpoints through without authentication. (optional, no default)

## Deployment

A very simple deployment example using [kpurdon/echosrv](https://github.com/kpurdon/echosrv) as the target service is shown below. Note that this does not include the [service/ingress level configuration](https://cloud.google.com/iap/docs/enabling-kubernetes-howto) needed by Cloud IAP.

``` yaml
apiVersion: v1
kind: Service
metadata:
  name: echosrv
  labels:
    app: echosrv
  annotations:
    beta.cloud.google.com/backend-config: '{"default": "echosrv}' # see cloud iap docs
spec:
  type: NodePort
  selector:
    app.kubernetes.io/name: echosrv
  ports:
  - protocol: TCP
    port: 80
    targetPort: 8000
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: echosrv
data:
  HTTP_PORT: "8001"
  IAPAP_PORT: "8000"
  IAPAP_TARGET: "http://localhost:8001"
  IAPAP_AUDIENCE: "/projects/12345/global/backendServices/12345" # replace me with your value
  IAPAP_ENDPOINT_WHITELIST: "/liveness,/readiness"
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: echosrv
  labels:
    app: echosrv
spec:
  replicas: 1
  selector:
    matchLabels:
      app.kubernetes.io/name: echosrv
  template:
    metadata:
      labels:
        app.kubernetes.io/name: echosrv
    spec:
      containers:
        - name: echosrv-iapap
          image: kpurdon/iapap:latest
          imagePullPolicy: IfNotPresent
          terminationMessagePolicy: FallbackToLogsOnError
          envFrom:
            - configMapRef:
                name: echosrv
          ports:
            - name: http
              containerPort: 8000
          livenessProbe:
            httpGet:
              port: http
              path: /_liveness
            periodSeconds: 30
          readinessProbe:
            httpGet:
              port: http
              path: /_readiness
            periodSeconds: 60
        - name: echosrv
          image: kpurdon/echosrv:latest
          imagePullPolicy: IfNotPresent
          terminationMessagePolicy: FallbackToLogsOnError
          envFrom:
            - configMapRef:
                name: echosrv
          ports:
            - name: http
              containerPort: 8001
          livenessProbe:
            httpGet:
              port: http
              path: /liveness
            periodSeconds: 30
          readinessProbe:
            httpGet:
              port: http
              path: /readiness
            periodSeconds: 60
```
