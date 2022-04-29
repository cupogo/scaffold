package stores

import "daxv.cn/gopak/lib/zlog"

func logger() zlog.Logger {
	return zlog.Get()
}
