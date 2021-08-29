all: wpscan

run:
	./scan --jobs 10000 -t 10 ./d/d2

wpscan:
	go build  -o scan ./...

clean:
	rm -rf errlog install.txt setup.txt

fclean: clean
	rm -rf scan

re: fclean all

.PHONY: all lclean fclean re wpscan
