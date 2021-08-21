all: wpscan

wpscan: scan.go main.go response.go
	go build -ldflags "-w -s" -o wpscan ./... 


lclean:
	echo > setup.txt > install.txt

fclean:
	rm -rf wpscan

re: fclean all

.PHONY: all lclean fclean re
