# Captain
Captain – the chess Swiss Pairing Engine

## Compile
```bash
make
```

## Run Pairing Engine
```bash
go run src/pairing_engine/main/main.go
```

## Run Test Interface
```bash
cd tools
python3 tournament_manager.py
```

**Important:** do not execute `python3 tools/tournament_manager.py` from the main directory because it'll create a subdirectory for test files in the wrong place.