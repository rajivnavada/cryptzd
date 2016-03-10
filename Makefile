CC_FOR_TARGET = clang
CXX_FOR_TARGET = clang++

all: gibberz

gibberz:
	CC_FOR_TARGET=$(CC_FOR_TARGET) CXX_FOR_TARGET=$(CXX_FOR_TARGET) go build -o gibberz gibberz.go


