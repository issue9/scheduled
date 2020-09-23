scheduled
[![Build Status](https://travis-ci.org/issue9/scheduled.svg?branch=master)](https://travis-ci.org/issue9/scheduled)
[![license](https://img.shields.io/badge/license-MIT-brightgreen.svg?style=flat)](https://opensource.org/licenses/MIT)
[![codecov](https://codecov.io/gh/issue9/scheduled/branch/master/graph/badge.svg)](https://codecov.io/gh/issue9/scheduled)
[![Go version](https://img.shields.io/badge/Go-1.13-brightgreen.svg?style=flat)](https://golang.org)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/issue9/scheduled)](https://pkg.go.dev/github.com/issue9/scheduled)
======

scheduled 是一个计划任务管理工具。

通过 scheduled 可以实现管理类似 linux 中 crontab 功能的计划任务功能。
当然功能并不止于此，用户可以实现自己的调度算法，定制任务的启动机制。

目前 scheduled 内置了以下三种算法：

- cron 实现了 crontab 中的大部分语法功能；
- at 在固定的时间点执行一次任务；
- ticker 以固定的时间段执行任务，与 time.Ticker 相同。

```go
srv := scheduled.NewServer(time.UTC)

ticker := func() error {
    _,err := fmt.Println("ticker @ ", time.Now())
    return err
}


expr := func() error {
    _,err := fmt.Println("cron @ ", time.Now())
    return err
}

srv.Tick(ticker, 1*time.Minute, false, false)
srv.Cron(expr, "@daily", false)
srv.Cron(expr, "* * 1 * * *", false)

log.Panic(srv.Serve())
```

版权
---

本项目采用 [MIT](https://opensource.org/licenses/MIT) 开源授权许可证，完整的授权说明可在 [LICENSE](LICENSE) 文件中找到。
