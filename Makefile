 # Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GORUN=$(GOCMD) run
GOSRCDIR=src
GOBUILDDIR=bin

# Set sources
SRCS := $(wildcard $(GOSRCDIR)/*.go)

all: test build

build:
		mkdir -p $(GOBUILDDIR)
		$(GOBUILD) -o $(GOBUILDDIR)/main $(SRCS)
run:
		$(GORUN) $(SRCS)

test: 
		$(GOTEST) -v ./...

clean: 
		rm $(GOBUILDDIR)/*