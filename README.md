# WP-ISS
High Performance Golang Wordpress spider based on [valyala/fasthttp](https://github.com/valyala/fasthttp).\
\
It may scan billion of sites in a hour.

**WARNING:**
  > Beware limit of maximum open files with:\
  <code>ulimit -n verylongnum </code>
  
# How to run

OPTIONS :
   - -j --jobs (num) number of goroutines to run (default 100)\
   - -l                    enables error log and bench\
   - -t (num)         dial timeout (default 3)\

<img src="/example/g.gif?raw=true">
