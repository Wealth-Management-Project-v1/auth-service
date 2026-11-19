#argocd_server = $(shell kubectl get svc argocd-server -n argocd -o jsonpath='{.status.loadBalancer.ingress[0].hostname}')
#argocd_password = $(shell kubectl get secret argocd-initial-admin-secret -n argocd -o jsonpath='{.data.password}' | base64 -d)

docker-build:
	git pull
	aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin 565122144984.dkr.ecr.us-east-1.amazonaws.com
	docker build -t 565122144984.dkr.ecr.us-east-1.amazonaws.com/auth-service:$(image_tag) .
	docker push 565122144984.dkr.ecr.us-east-1.amazonaws.com/auth-service:$(image_tag)

eks-deploy:
	git pull
	aws eks update-kubeconfig --name wmp-dev
	helm upgrade -i auth-service helm -f helm/values/auth-service.yaml --set image_tag=$(image_tag)

argocd-deploy:

	aws eks update-kubeconfig --region us-east-1 --name wmp-dev

	@echo "Checking Kubernetes connection..."
	kubectl get nodes

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