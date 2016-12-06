# etcdd

etcd service discovery, compatible with v2 and v3 etcd.

## getting

```sh
go get github.com/4396/etcdd
```

## usage

### new discoverer

```golang
endpoints := []string{
    "http://127.0.0.1:2379",
}

dv2, err := etcdd.NewV2(endpoints)
if err != nil {
    // ...
}
dv3, err := etcdd.NewV3(endpoints)
if err != nil {
    // ...
}
```

### register and unregister

```golang
svrname := "svr1"
svraddr := "127.0.0.1:1234"
namespace := "/services"
timeout := time.Second * 5

// register service, returns keepalive function
keepalive, err := dv3.Register(namespace, svrname, svraddr, timeout * 2)
if err != nil {
    // ...
}
// unregister service
defer dv3.Unregister(namespace, svrname)

// keep alive service
ticker := time.NewTicker(timeout)
for {
    select {
    case <-ticker.C:
        if err = keepalive(); err != nil {
            // ...
        }
    }
}
```

### watch namespace

```golang
// watch namespace, returns chan event and cancel function
event, cancel, err := dv3.Watch(namespace)
if err != nil {
    // ...
}
// cancel watch
defer cancel()

for {
    select {
    case ev, ok := <-event:
        if ok {
            fmt.Println(ev.Action, ev.Name, ev.Addr)
        }
    }
}
```

### list services

```golang
svrs, err := dv3.Services(namespace)
if err != nil {
    // ...
}
for name, addr := range svrs {
    fmt.Println(name, addr)
}
```

### version

```golang
ver := dv3.Version()
if ver != 3 {
    // ...
}
```

### close

```golang
err := dv3.Close()
if err != nil {
    // ...
}
```

## example

- [api usage](./example/etcdd)
- [a simple chat room](./example/chat)
