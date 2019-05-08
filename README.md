scheduled
[![Build Status](https://travis-ci.org/issue9/scheduled.svg?branch=master)](https://travis-ci.org/issue9/scheduled)
[![license](https://img.shields.io/badge/license-MIT-brightgreen.svg?style=flat)](https://opensource.org/licenses/MIT)
[![codecov](https://codecov.io/gh/issue9/scheduled/branch/master/graph/badge.svg)](https://codecov.io/gh/issue9/scheduled)
[![Go version](https://img.shields.io/badge/Go-1.5-brightgreen.svg?style=flat)](https://golang.org)
[![GoDoc](https://godoc.org/github.com/issue9/scheduled?status.svg)](https://godoc.org/github.com/issue9/scheduled)
======

定时任务管理工具。
```go
srv := scheduled.NewServer()

ticker := func() error {
    _,err := fmt.Println("ticker @ ", time.Now())
    return err
}


expr := func() error {
    _,err := fmt.Println("expr @ ", time.Now())
    return err
}

srv.NewTicker(ticker, 1*time.Minute)
srv.NewCron(expr, "@daily")
srv.NewCron(expr, "* * 1 * * *")

log.Panic(srv.Serve())
```


### 版权

本项目采用 [MIT](https://opensource.org/licenses/MIT) 开源授权许可证，完整的授权说明可在 [LICENSE](LICENSE) 文件中找到。

