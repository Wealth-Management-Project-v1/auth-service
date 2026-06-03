docker-build:
	git pull
	aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin 200915316788.dkr.ecr.us-east-1.amazonaws.com
	docker build -t  200915316788.dkr.ecr.us-east-1.amazonaws.com/auth-service:latest .
	docker push 200915316788.dkr.ecr.us-east-1.amazonaws.com/auth-service:latest

eks-deploy:
	git pull
	aws eks update-kubeconfig --name wmp-dev
	helm upgrade -i auth-service helm -f helm/values/auth-service --set image_tag=$(image_tag)