BUILD_DIR=build
IOS_ARTIFACT=$(BUILD_DIR)/SshLib.xcframework

LDFLAGS="-s -w"
IMPORT_PATH=sshlib

goDeps:
	go get -d ./...
	go get golang.org/x/mobile/cmd/gomobile
	gomobile init
	# go get -u github.com/golang/protobuf/protoc-gen-go

init_env: clean goDeps 
	@echo DONE

build_apple:
	go get golang.org/x/mobile/cmd/gomobile
	gomobile init

	mkdir -p $(BUILD_DIR)
	gomobile bind -a -v -ldflags $(LDFLAGS) -target=ios,iossimulator,macos -o $(IOS_ARTIFACT) $(IMPORT_PATH)

clean:
	rm -rf $(BUILD_DIR)
