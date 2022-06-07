# cache-go
****
该项目实现了一个分布式缓存系统。使用TCP作为通信协议，ProtoBuf作为序列化与反序列协议。总体采用Reactor模式，内置连接池、
负载均衡、缓存淘汰机制LRU、LFU等机制，同时采用一致性哈希将数据在不同节点间进行分片。


|  作者  |                                 刘庆夫                                  |
|:----:|:--------------------------------------------------------------------:|
| CSDN | [我的CSDN](https://blog.csdn.net/sbsbsbsbsb_1?spm=1000.2115.3001.5343) |
| 掘金主页 |           [我的掘金](https://juejin.cn/user/1236291634601191)            |
****
## 目录
* [install](#install)
* [使用与配置](#使用与配置)
  * 新建server
  * 集群配置
  * 启动
  * 添加缓存节点
* [技术点](#install)
  * [LRU与LFU](#LRU与LFU)
  * [防缓存击穿](#防缓存击穿)
  * [连接池](#连接池)
  * [Reactor](#Reactor)
  
## install
****
```
go get github.com/qingfuliu/cache-go.git
```

使用与配置
------
****
##新建server：
```go
p, err := NewTcpCacheServer(proto, Addr string,options ...Options)
```
可选配置项：

| 名称           | 设置                   |
|--------------|----------------------|
| proto,addr   | 协议与地址                |
| CodeC        | tcp编码解码器             |
| LoadBalancer | 负载均衡策略               |
| Socket       | server Socker文件描述符设置 |


## 集群配置：

```go
err := p.AddRemoteAddr("tcp", ":5201")
```
## 启动服务
```go
err:=p.Start(lockOsThread bool, numReactor int, option ...TcpPoolOption)
```
可选配置项

| 名称           | 设置                         |
|--------------|----------------------------|
| lockOsThread        | 是否开启runtime.LockOSThread() |
| numReactor | reactor节点个数                |
| TcpPoolOption       | 协程池相关参数                    |
##添加缓存节点

```go
 cache_go.NewCacheHub(name string, getter Getter, maxBytes int64, options ...CacheHubOption)
```

| 名称           | 设置               |
|--------------|------------------|
| name        | 节点名称             |
| getter | set的方式，必须设置      |
| SetLruCache()       | 缓存淘汰机制为LRU（默认配置） |
|SetLfuCache()| 缓存淘汰机制为LFU       |

技术点
------
****
### 1.LRU与LFU
* LRU:
  * 使用链表+hash表构建，缓存查找和淘汰都有O（1）时间复杂度

* LFU
  * 使用RB_TREE+hash表构建，RB_TREE维护一个最左节点，保证在缓存淘汰时有O（1）的事件复杂度
### 2.连接池
整体借鉴mysql数据库驱动包的连接池。可选配置：

|                     名称                     |       作用        |
|:------------------------------------------:|:---------------:|
|          SetMaxIdle(maxIdle int)           |  设置最大的idel连接数   |
| SetMaxIdleTime(maxIdleTime time.Duration)  |   设置最大的idle时间   |
| SetMaxOpen(maxOpen int64)|  设置最大能够打开的连接数   |
| SetMaxLife(maxLife time.Duration)|   设置连接最大存活时间    |
| setCodeC(codec CodeC)| 设置解码编码器，用于粘包等场景 |
| setLocalAddr(localAddr net.Addr)|     设置本机地址      |
### 3.防缓存击穿
关键代码在ConcurrencyBarrier.go(并发屏障)文件下，解释如下：
```go
        //整体思路：
        //对于每一个并发连接，记录其所检索的key
        //如果是第一个到达的连接，将改key放入map中，并将返回值地址提前存储，之后开始执行，执行完毕后唤醒阻塞的并发连接
        //如果是后到达的连接，直接阻塞在map，等待第一个连接返回结果即可

	cb.mu.Lock()
	if cl, ok := cb.m[key]; ok {
		//说明不是新key
		cb.mu.Unlock()
		//等待第一个连接返回 
		cl.wg.Wait()
		return cl.val, cl.err
	}
	//将结果存储在cl中
	cl := &call{}
	cl.wg.Add(1)
	cb.m[key] = cl
	cb.mu.Unlock()
	//调用函数
	cl.val, cl.err = fn()

	cl.wg.Done()
	cb.mu.Lock()
	delete(cb.m, key)
	cb.mu.Unlock()
	return cl.val, cl.err
```
###4.Reactor



[回到顶点](#cache-go)
