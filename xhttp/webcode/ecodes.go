package webcode

// commom ecode
var (
	StatusOK = add(200, "StatusOK")

	RequestErr   = add(4000, "请求错误")
	NothingFound = add(4001, "路由错误")
	ParamInvalid = add(4002, "参数错误")
	Unauthorized = add(4100, "未认证")
	TokenExpires = add(4101, "认证过期")
	AccessDenied = add(4102, "权限不足")

	ServerErr          = add(5000, "服务器错误")
	ServiceUnreachable = add(5001, "服务暂不可用")
	ServiceTimeout     = add(5002, "服务调用超时")
	DatabaseErr        = add(5100, "数据库操作失败")
	DataNotFound       = add(5101, "未查询到数据")
	InternalErr        = add(5200, "内部请求错误")
)
