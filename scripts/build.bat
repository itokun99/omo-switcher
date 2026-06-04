@echo off
echo Building omo-switch...

go build -o omo-switch.exe ./cmd/omo-switch

echo Build complete: omo-switch.exe
