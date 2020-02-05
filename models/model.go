package models

import (
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
	"time"
)

// 一共有3张表
// 用户表
type User struct {
	Name     string `orm:"pk"`
	Pwd      string
	Articles [] *Article `orm:"reverse(many)"` // 设置一对多关系，一个用户可以发表多篇文章
}

// 文章表
type Article struct {
	// 注意字段名一般为首字母大写；反引号+（主键pk，自增auto）
	Id    int    `orm:"pk;auto"`
	Aname string `orm:"size(20)"`
	// auto_now 每次model保存时都会对时间自动更新
	Atime    time.Time `orm:"auto_now"`
	Acount   int       `orm:"default(0);null"`
	Acontent string // 文章内容
	Aimg     string // 文章存放路径
	Atype    *ArticleType `orm:"rel(fk)"`       //设置一对多关系，一篇文章对应一个类型，文章类型为外键
	User    *User     `orm:"rel(fk)"` //设置一对多的关系，一篇文章对应一个创建者
}

// 文章类型表
type ArticleType struct {
	Id       int
	Typename string      `orm:"size(20)"`
	Articles [] *Article `orm:"reverse(many)"` // 设置一对多的反向关系，一个类型可以包含多篇文章
}

func init() {
	// 设置数据库基本信息
	_ = orm.RegisterDataBase("default", "mysql", "登录账号:登录密码@tcp(mysql主机ip地址:3306)/数据库名?charset=utf8")
	// 注册模型
	orm.RegisterModel(new(User), new(Article), new(ArticleType))
	// 生成表,第一个参数是别名，第2个参数是否强制更新（除非表结构发生改变），一般设置为false，若为true，则表数据会被清空！
	// 第3个参数是是否打印日志（打印sql语句）
	_ = orm.RunSyncdb("default", false, true)
}
