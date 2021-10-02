
set NDK=C:\Users\Stein\ndk-bundle\toolchains\llvm\prebuilt\windows-x86_64\bin
set CC=%NDK%\aarch64-linux-android21-clang
set GOOS=android
set CGO_ENABLED=1
set GOARCH=arm64
set GOARM=7

go build -ldflags "-s -w" -o simple-injector-android64
