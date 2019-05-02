.PHONY: build-debug build-staging test


# Escape ':' character if present in command that needs to be executed from makefile

build-debug:
	./gradlew sdk\:test && ./gradlew sdk\:assembleDebug

build-staging:
	./gradlew sdk\:test && ./gradlew sdk\:assembleStaging


test:
	./gradlew sdk\:test
