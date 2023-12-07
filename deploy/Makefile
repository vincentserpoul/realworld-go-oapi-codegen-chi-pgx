.PHONY: api/deploy-localdev migrate/deploy-localdev api/kustomize-manifest-gen migrate/kustomize-manifest-gen context-select

ENV=localdev
NAMESPACE=realworkd

api/deploy-localdev:
	kubectl kustomize ./$(API_NAME)/$(ENV) |  kubectl apply -f -
	kubectl -n $(NAMESPACE) rollout restart deployment $(API_NAME)

migrate/deploy-localdev:
	kubectl kustomize ./$(MIGRATE_NAME)/$(ENV) |  kubectl apply -f -
	kubectl -n $(NAMESPACE) rollout restart statefulset $(MIGRATE_NAME)

api/kustomize-manifest-gen:
	kubectl kustomize ./$(API_NAME)/$(ENV)

migrate/kustomize-manifest-gen:
	kubectl kustomize ./$(MIGRATE_NAME)/$(ENV)

##########
# Others #
##########

context-select:
	kubectl config use-context $(K8S_CONTEXT)
