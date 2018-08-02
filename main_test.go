package main

import (
    "testing"
)

func Test_app(t *testing.T) {
    logger.Debug("测试parseArgs")
    app := NewApp()
    app.IsTestCase = true
    app.Run()
}