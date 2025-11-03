package auth

// LoginReq 用户登录请求参数
type LoginReq struct {
	Username string `json:"username" validate:"required" label:"用户名"`
	Password string `json:"password" validate:"required" label:"密码"`
}

// LoginRes 用户登录响应
type LoginRes struct {
	AccessToken  string `json:"access_token" label:"访问令牌"`
	RefreshToken string `json:"refresh_token" label:"刷新令牌"`
	ExpiresIn    int64  `json:"expires_in" label:"访问令牌过期秒数"`
}

// RefreshTokenReq 刷新令牌请求
type RefreshTokenReq struct {
	RefreshToken string `json:"refresh_token" validate:"required" label:"刷新令牌"`
}

// RefreshTokenRes 刷新令牌响应
type RefreshTokenRes struct {
	AccessToken  string `json:"access_token" label:"访问令牌"`
	RefreshToken string `json:"refresh_token" label:"刷新令牌"`
	ExpiresIn    int64  `json:"expires_in" label:"访问令牌过期秒数"`
}

// UserDetailRes 用户详情响应
type UserDetailRes struct {
	ID          string        `json:"id" label:"用户ID"`
	Username    string        `json:"username" label:"用户名"`
	Phone       string        `json:"phone,omitempty" label:"手机号"`
	Profile     interface{}   `json:"profile,omitempty" label:"个人信息"`
	Roles       []UserRoleRes `json:"roles" label:"角色列表"`
	Permissions []UserPermRes `json:"permissions" label:"权限列表"`
	CreatedAt   string        `json:"created_at" label:"创建时间"`
	UpdatedAt   string        `json:"updated_at" label:"更新时间"`
}

// UserRoleRes 用户角色条目
type UserRoleRes struct {
	ID          string  `json:"id" label:"角色ID"`
	Name        string  `json:"name" label:"角色名称"`
	Description *string `json:"description,omitempty" label:"角色描述"`
}

// UserPermRes 用户权限条目
type UserPermRes struct {
	ID     string `json:"id" label:"权限ID"`
	Name   string `json:"name" label:"权限名称"`
	Type   string `json:"type" label:"权限类型"`
	Code   string `json:"code,omitempty" label:"权限编码"`
	Path   string `json:"path,omitempty" label:"接口路径"`
	Method string `json:"method,omitempty" label:"请求方法"`
}
