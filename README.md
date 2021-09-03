# WP-ISS
High Performance Golang Wordpress spider based on [valyala/fasthttp](https://github.com/valyala/fasthttp).\
\
It may run more than 1000 scanner goroutines at once.

**WARNING:**
  > Beware limit of maximum open files with:\
  <code>ulimit -n verylongnum </code>
  
# How to run

OPTIONS :
   - -j --jobs (num) number of goroutines to run (default 100)\
   - -l                    enables error log and bench\
   - -t (num)         dial timeout (default 3)\

<img src="/assets/g.gif?raw=true">
