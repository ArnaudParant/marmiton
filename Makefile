PROJECT := marmiton
export BUILD_TAG := $(USER)


docker:			Dockerfile
	docker build -f Dockerfile --network host -t "$(PROJECT):$(BUILD_TAG)" .
