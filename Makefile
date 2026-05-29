.PHONY: test run view measure

test:
	go test ./...

measure:
	go test ./pkg/sketch/ -run TestSketch_PrintMeasurements -v

run:
	go run ./cmd

view:
	go test ./pkg/sketch/... && go run ./cmd/viewer ./pkg/sketch/testdata