all: wpscan

wpscan: scan.go
	go build -ldflags "-w -s" -o wpscan scan.go


lclean:
	echo > setup.txt > install.txt

fclean:
	rm -rf wpscan

re: fclean all

.PHONY: all lclean fclean re
