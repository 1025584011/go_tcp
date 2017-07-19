# go_tcp
a tcp server  implement by epoll


用GO实现的tcp server，可以自定义协议，目前用简单http协议测试，性能超越 GO自身的 http库，下面是ab测试结果


go net/http

[root@compile ~]# ab -n 1000000 -c 1000 http://127.0.0.1:11112/

Concurrency Level:      1000
Time taken for tests:   28.828 seconds
Complete requests:      1000000
Failed requests:        0
Write errors:           0
Total transferred:      120000000 bytes
HTML transferred:       4000000 bytes
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

Percentage of the requests served within a certain time (ms)
  50%     28
  66%     28
  75%     29
  80%     29
  90%     31
  95%     32
  98%     33
  99%     35
 100%   3038 (longest request)




go_tcp 

[root@compile ~]# ab -n 1000000 -c 1000 http://127.0.0.1:11111/

Document Path:          /
Document Length:        4 bytes

Concurrency Level:      1000
Time taken for tests:   24.211 seconds
Complete requests:      1000000
Failed requests:        0
Write errors:           0
Total transferred:      120000000 bytes
HTML transferred:       4000000 bytes
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

Percentage of the requests served within a certain time (ms)
  50%      4
  66%      5
  75%      6
  80%      7
  90%     10
  95%     13
  98%     25
  99%   1006
 100%   7022 (longest request)
