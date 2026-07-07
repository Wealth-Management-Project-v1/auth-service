argocd_login = kubectl get svc argo-argocd-server -o jsonpath="{.status.loadBalancer.ingress[].hostname}"
argocd_password = kubectl get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d


docker-build:
	git pull
	aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin 213026892552.dkr.ecr.us-east-1.amazonaws.com
	docker build -t  213026892552.dkr.ecr.us-east-1.amazonaws.com/auth-service:$(image_tag) .
	docker push 213026892552.dkr.ecr.us-east-1.amazonaws.com/auth-service:$(image_tag)

eks-deploy:
	git pull
	aws eks update-kubeconfig --name wmp-dev
	helm upgrade -i auth-service helm -f helm/values/auth-service.yaml --set image_tag=$(image_tag)

argo-deploy:
	git pull
	aws eks update-kubeconfig --name wmp-dev

	argocd login  $(argocd_login) \
  	--insecure \
  	--username admin \
  	--password $(argocd_password)

	argocd app create auth-service \
    --repo https://github.com/Wealth-Management-Project-v1/helm-v1.git \
    --path . \
    --revision main \
    --dest-server https://kubernetes.default.svc \
    --dest-namespace default \
    --sync-policy auto \
    --values values/auth-service.yaml \
    --helm-set-string $(image_tag)


