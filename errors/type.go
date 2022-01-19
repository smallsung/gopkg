package errors

type Caller interface {
	// Caller 返回包装错误调用栈
	Caller() uintptr
}
type Callers interface {
	// Callers 返回包装错误完整调用栈
	Callers() []uintptr
}

type Location interface {
	// Location 返回调用包装错误的文件和行
	Location() (string, int)
}

type Annotator interface {
	//Annotate 返回错误注释信息
	Annotate() string
}

type UnWrapper interface {
	//Unwrap go1.13支持. 不建议返回 nil
	Unwrap() error
}

type Locator interface {
	//SetLocation 设置调用包装的文件和行.供 locator 调用
	SetLocation(skip int)
}
