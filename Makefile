GO?=go
GOBIN?=~/go/bin
CP?=cp

EXECS=broker/client/broker broker/daemon/brokerd deploy/deploy perform/perform

.PHONY: clean providers brokerd broker system deploy perform utils clean deps $(EXECS)

all: utils providers system broker cluster deploy perform

utils:
	@(cd utils && $(MAKE))

providers:
	@(cd providers && $(MAKE))

system:
	@(cd system && $(MAKE))

broker: utils system providers
	@(cd broker && $(MAKE))

cluster: utils system providers broker
	@(cd cluster && $(MAKE))

deploy: utils system providers broker cluster
	@(cd deploy && $(MAKE))

perform: utils system providers broker
	@(cd perform && $(MAKE))

clean:
	@(cd providers && $(MAKE) $@)
	@(cd system && $(MAKE) $@)
	@(cd broker && $(MAKE) $@)
	@(cd cluster && $(MAKE) $@)
	@(cd deploy && $(MAKE) $@)
	@(cd perform && $(MAKE) $@)

broker/client/broker: broker

broker/daemon/brokerd: broker

deploy/deploy: deploy

perform/perform: perform

install: $(EXECS)
	@($(CP) -f $^ $(GOBIN))

# DEPENDENCIES MANAGEMENT
#Generate enum tring
STRINGER := golang.org/x/tools/cmd/stringer
#Embed shell file into code go
RICE := github.com/GeertJohan/go.rice github.com/GeertJohan/go.rice/rice
#CLI parser
URFAVE := github.com/urfave/cli
MOW := github.com/jawher/mow.cli
#Configuration file handler
VIPER := github.com/spf13/viper
#Data validation lib: at least used to validate host name for flexibleengine
PENGUS_CHECK := github.com/pengux/check
UUID := github.com/satori/go.uuid
SPEW := github.com/davecgh/go-spew/spew
DSP := github.com/mjibson/go-dsp/fft
TESTIFY := github.com/stretchr/testify
PASSWORD := github.com/sethvargo/go-password/password
CRYPTO_SSH := golang.org/x/crypto/ssh

# GRPC LIBS
CONTEXT := golang.org/x/net/context
GRPC := google.golang.org/grpc
PROTOBUF := github.com/golang/protobuf/protoc-gen-go
GRPC_LIBS := $(GRPC) $(PROTOBUF) $(CONTEXT)

## Providers SDK
# OpenStack SDK for GO
GOPHERCLOUD := github.com/gophercloud/gophercloud
# AWS SDK for GO
AWS := github.com/aws/aws-sdk-go
# Providers SDK
PROVIDERS_SDK := $(GOPHERCLOUD) $(AWS)

DEPS := $(STRINGER) $(RICE) $(URFAVE) $(VIPER) $(PENGUS_CHECK) $(UUID) $(SPEW) $(DSP) $(TESTIFY) $(PASSWORD) $(CRYPTO_SSH) $(GRPC_LIBS) $(PROVIDERS_SDK) $(MOW)

deps: ; $(GO) get -u $(DEPS)
