.PHONY: preflight up down status test wipe

preflight:
	./preflight.sh

up:
	./up.sh

down:
	./down.sh

wipe:
	./down.sh --wipe

status:
	./status.sh

test:
	go test -tags e2e -count=1 ./
