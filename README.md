# go_tcp
a tcp server  implement by epoll


用GO实现的tcp server，可以自定义协议，目前用简单http协议测试，性能超越 GO自身的 http库，下面是ab测试结果


go net/http

[root@compile ~]# ab -n 1000000 -c 1000 http://127.0.0.1:11112/                                                                                                                 

Requests per second:    34688.15 [#/sec] (mean)

Time per request:       28.828 [ms] (mean)

Time per request:       0.029 [ms] (mean, across all concurrent requests)

Transfer rate:          4065.02 [Kbytes/sec] received

Connection Times (ms)
              min  mean[+/-sd] median   max
Connect:        0   14  43.5     13    3019

Processing:     2   14   2.8     14      36

Waiting:        0   10   3.0      9      32

Total:          9   29  43.5     28    3038



go_tcp 

[root@compile ~]# ab -n 1000000 -c 1000 http://127.0.0.1:11111/

Requests per second:    41303.95 [#/sec] (mean)

Time per request:       24.211 [ms] (mean)

Time per request:       0.024 [ms] (mean, across all concurrent requests)

Transfer rate:          4840.31 [Kbytes/sec] received

Connection Times (ms)
              min  mean[+/-sd] median   max
Connect:        0   19 150.3      0    7018
Processing:     1    4  11.8      3    1604
Waiting:        0    4  11.8      3    1604
Total:          1   23 153.7      4    7022


