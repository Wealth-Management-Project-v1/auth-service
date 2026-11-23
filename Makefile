
AWS_ACCOUNT_ID := 565122144984
AWS_REGION := us-east-1
ECR_REGISTRY := $(AWS_ACCOUNT_ID).dkr.ecr.$(AWS_REGION).amazonaws.com
IMAGE_NAME := analytics-service
IMAGE_URI := $(ECR_REGISTRY)/$(IMAGE_NAME):$(image_tag)

.PHONY: docker-build eks-deploy argocd-deploy

docker-build:
	@test -n "$(image_tag)" || (echo "ERROR: image_tag is required"; exit 1)
	aws ecr get-login-password --region $(AWS_REGION) | docker login --username AWS --password-stdin $(ECR_REGISTRY)
	docker build -t $(IMAGE_URI) .
	docker push $(IMAGE_URI)

eks-deploy:
	@test -n "$(image_tag)" || (echo "ERROR: image_tag is required"; exit 1)
	aws eks update-kubeconfig --region $(AWS_REGION) --name wmp-dev
	helm upgrade --install analytics-service helm \
		-f helm/values/analytics-service.yaml \
		--set-string image_tag=$(image_tag)

argocd-deploy:
	@test -n "$(image_tag)" || (echo "ERROR: image_tag is required"; exit 1)
	aws eks update-kubeconfig --region $(AWS_REGION) --name wmp-dev
	@ARGOCD_SERVER=$$(kubectl get svc argocd-server -n argocd -o jsonpath='{.status.loadBalancer.ingress[0].hostname}'); \
	ARGOCD_PASSWORD=$$(kubectl get secret argocd-initial-admin-secret -n argocd -o jsonpath='{.data.password}' | base64 -d); \
	if [ -z "$$ARGOCD_SERVER" ]; then echo "ERROR: ArgoCD server hostname is empty"; exit 1; fi; \
	if [ -z "$$ARGOCD_PASSWORD" ]; then echo "ERROR: ArgoCD password is empty"; exit 1; fi; \
	echo "ArgoCD server hostname retrieved successfully"; \
	argocd login "$$ARGOCD_SERVER" --insecure --username admin --password "$$ARGOCD_PASSWORD"

	argocd app create analytics-service \
		--repo https://github.com/Wealth-Management-Project-v1/helm-v1.git \
		--path . \
		--revision main \
		--dest-server https://kubernetes.default.svc \
		--dest-namespace default \
		--sync-policy auto \
		--upsert \
		--values values/analytics-service.yaml \
		--helm-set-string image_tag=$(image_tag)

	argocd app sync analytics-service
	argocd app wait analytics-service --health --timeout 300

