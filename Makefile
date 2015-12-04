PUBLISH=publish_docker_env

.DEFAULT: all
.PHONY: all update tests publish $(PUBLISH) clean prerequisites build travis run-smoketests

# If you can use docker without being root, you can do "make SUDO="
SUDO=sudo

DOCKERHUB_USER=inercia
DOCKER_ENV_VERSION=git-$(shell git rev-parse --short=12 HEAD)

DOCKER_ENV_EXE=prog/docker-env
EXES=$(DOCKER_ENV_EXE)
DOCKER_ENV_UPTODATE=.docker_env.uptodate
IMAGES_UPTODATE=$(DOCKER_ENV_UPTODATE)
DOCKER_ENV_IMAGE=$(DOCKERHUB_USER)/docker-env
IMAGES=$(DOCKER_ENV_IMAGE)
DOCKER_ENV_EXPORT=docker_env.tar

all:    $(EXES)
travis: $(EXES)

update:
	go get -u -f -v $(addprefix ./,$(dir $(EXES)))

$(DOCKER_ENV_EXE):
	go get -tags netgo ./$(@D)
	go build -ldflags "-extldflags \"-static\" -X main.version $(DOCKER_ENV_VERSION)" -o $@ ./$(@D)

$(DOCKER_ENV_EXE): Makefile prog/*.go prog/*/*.go env/*.go env/*/*.go

$(DOCKER_ENV_UPTODATE): Dockerfile
	$(SUDO) docker build -t $(DOCKER_ENV_IMAGE) .
	touch $@

$(DOCKER_ENV_EXPORT): $(IMAGES_UPTODATE)
	$(SUDO) docker save $(addsuffix :latest,$(IMAGES)) > $@

$(PUBLISH): publish_%:
	$(SUDO) docker tag -f $(DOCKERHUB_USER)/$* $(DOCKERHUB_USER)/$*:$(DOCKER_ENV_VERSION)
	$(SUDO) docker push   $(DOCKERHUB_USER)/$*:$(DOCKER_ENV_VERSION)
	$(SUDO) docker push   $(DOCKERHUB_USER)/$*:latest

publish: $(PUBLISH)

image: clean-local $(DOCKER_ENV_UPTODATE)

clean-local:
	rm -f $(EXES) $(IMAGES_UPTODATE) $(DOCKER_ENV_EXPORT) test/tls/*.pem coverage.html profile.cov

clean: clean-local
	-$(SUDO) docker rmi $(IMAGES) 2>/dev/null

deps:
	go get ./...

dist: $(DOCKER_ENV_EXPORT)

test:
	go test ./...
