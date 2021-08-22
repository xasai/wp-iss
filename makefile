all: wpscan

wpscan:
	go build -ldflags "-w -s" -o scan ./... 

clean:
	echo > setup.txt > install.txt

fclean: clean
	rm -rf scan

re: fclean all

.PHONY: all lclean fclean re wpscan
