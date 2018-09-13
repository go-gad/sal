package foo

type Body struct{}

func (b Body) Query() string { return `` }

type List []*Body
