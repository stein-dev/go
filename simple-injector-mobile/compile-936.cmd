
set NDK=C:\Users\Stein\Downloads\Compressed\android-ndk-r17c-windows-x86_64\android-ndk-r17c\toolchains\arm-linux-androideabi-4.9\prebuilt\windows-x86_64\bin
set CC=%NDK%\arm-linux-androideabi-gcc
set GOOS=linux
set CGO_ENABLED=0
set GOARCH=arm
set GOARM=5

go build -ldflags "-s -w" -o si9
