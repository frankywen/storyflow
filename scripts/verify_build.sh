#!/bin/bash

# StoryFlow 构建验证脚本

echo "==================================="
echo "StoryFlow Build Verification"
echo "==================================="
echo ""

# 检查后端
echo ">>> Checking Backend..."
cd backend

echo "    - Go modules..."
go mod tidy
if [ $? -eq 0 ]; then
    echo "    ✓ Go modules OK"
else
    echo "    ✗ Go modules FAILED"
    exit 1
fi

echo "    - Go build..."
go build -o bin/server ./cmd/server
if [ $? -eq 0 ]; then
    echo "    ✓ Go build OK"
else
    echo "    ✗ Go build FAILED"
    exit 1
fi

cd ..

# 检查前端
echo ""
echo ">>> Checking Frontend..."
cd frontend

echo "    - TypeScript..."
npx tsc --noEmit
if [ $? -eq 0 ]; then
    echo "    ✓ TypeScript OK"
else
    echo "    ✗ TypeScript FAILED"
    exit 1
fi

echo "    - Build..."
npm run build
if [ $? -eq 0 ]; then
    echo "    ✓ Build OK"
else
    echo "    ✗ Build FAILED"
    exit 1
fi

cd ..

echo ""
echo "==================================="
echo "✓ All checks passed!"
echo "==================================="
echo ""
echo "To start the application:"
echo "  1. cp .env.example .env  # Configure your API keys"
echo "  2. docker-compose up -d postgres  # Start database"
echo "  3. cd backend && go run ./cmd/server  # Start backend"
echo "  4. cd frontend && npm run dev  # Start frontend"