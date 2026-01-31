CXX      ?= g++
CXXFLAGS := -std=c++17 -Wall -Wextra -g -I./include
LDFLAGS  := -lcurl

# 目录定义
OUT      ?= bin
OBJ_DIR  := obj
TARGET	 := $(OUT)/send

SRC_DIR  := src
# 源文件定义
SRCS     := $(wildcard $(SRC_DIR)/*.cpp)
OBJS 	 := $(SRCS:$(SRC_DIR)/%.cpp=$(OBJ_DIR)/%.o)

all: $(TARGET)

$(TARGET): $(OBJS) | $(OUT)
	$(CXX) $(OBJS) -o $(TARGET) $(LDFLAGS)
	@echo compile finished!

$(OBJ_DIR)/%.o: $(SRC_DIR)/%.cpp | $(OBJ_DIR)
	$(CXX) $(CXXFLAGS) -c $< -o $@

$(OUT) $(OBJ_DIR):
	mkdir -p $@

clean:
	rm -rf $(OBJ_DIR) $(OUT)
	@echo "Finish clean."

.PHONY: all zip clean

