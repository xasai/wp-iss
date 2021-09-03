all: build 

test: build
	./wpscan -l test
build:
	go build -o wpscan ./...

clean:
	rm -rf result

fclean: clean
	rm -rf scan

re: fclean all

.PHONY: all lclean fclean re wpscan test
