VERSION ?= 8
JAEGER_ENDPOINT ?= http://jaeger-collector.jaeger:14268/api/traces
DOCKER_NAMESPACE ?= pavolloffay

.PHONY: docker
docker:
	docker build -t $(DOCKER_NAMESPACE)/knative-tracing-first:$(VERSION) -f ./cmd/first/Dockerfile .
	docker build -t $(DOCKER_NAMESPACE)/knative-tracing-second:$(VERSION) -f ./cmd/second/Dockerfile .

.PHONY: docker-push
docker-push:
	docker push $(DOCKER_NAMESPACE)/knative-tracing-first:$(VERSION)
	docker push $(DOCKER_NAMESPACE)/knative-tracing-second:$(VERSION)

.PHONY: deploy
deploy:
	kn service create first \
    --image $(DOCKER_NAMESPACE)/knative-tracing-first:$(VERSION) \
    --port 8090 \
    --env OTEL_EXPORTER_JAEGER_ENDPOINT=$(JAEGER_ENDPOINT) \
    --env ENV_BROKER_URL=http://broker-ingress.knative-eventing.svc.cluster.local/default/knative-tracing \
    --revision-name=1
	kn service create second \
    --image $(DOCKER_NAMESPACE)/knative-tracing-second:$(VERSION) \
    --port 8080 \
    --env OTEL_EXPORTER_JAEGER_ENDPOINT=$(JAEGER_ENDPOINT) \
    --revision-name=1
	kubectl apply -f deploy/02-broker.yaml
	kn trigger create first-to-second --sink second --broker knative-tracing

clean:
	kn service delete first
	kn service delete second
	kn trigger delete first-to-second
	kubectl delete -f deploy/02-broker.yaml
