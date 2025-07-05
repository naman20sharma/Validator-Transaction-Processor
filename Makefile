.PHONY: build run test clean server full reset

build:
	@bash scripts/build.sh

run:
	@bash scripts/run.sh

test:
	@bash scripts/test.sh

server:
	@bash scripts/server.sh

clean:
	rm -rf bin/validator

reset:
	@echo "Cleaning up..."
	@pkill -f "bin/validator" || true
	@pkill -f "server.go" || true
	@lsof -ti udp:2001 | xargs kill -9 || true
	@lsof -ti tcp:2002 | xargs kill -9 || true
	@rm -rf bin/validator
	@echo "Reset complete. Ready for fresh start."

full:
	@echo "Building validator..."
	@bash scripts/build.sh
	@echo "Starting HTTP server..."
	@if [ "$$(uname)" = "Darwin" ]; then \
		osascript -e 'tell app "Terminal" to do script "cd $(PWD) && bash scripts/server.sh"'; \
	else \
		gnome-terminal -- bash -c "cd $(PWD) && bash scripts/server.sh; exec bash"; \
	fi
	@sleep 1
	@echo "Starting validator..."
	@if [ "$$(uname)" = "Darwin" ]; then \
		osascript -e 'tell app "Terminal" to do script "cd $(PWD) && bash scripts/run.sh"'; \
	else \
		gnome-terminal -- bash -c "cd $(PWD) && bash scripts/run.sh; exec bash"; \
	fi
	@sleep 1
	@echo "Validator and server started. Run 'make test' when ready to send transactions."
	