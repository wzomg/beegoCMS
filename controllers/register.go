package controllers

import (
	"beegoDemo/models"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
)

type RegisterController struct {
	beego.Controller
}

func (c *RegisterController) TurnToRegister() {
	c.Data["errmsg"] = ""
	c.TplName = "register.html"
}

func (c *RegisterController) HandleRegister() {
	// 1、拿到url数据
	userName := c.GetString("userName")
	pwd := c.GetString("pwd")
	// 2、校验数据并插入数据库
	o := orm.NewOrm()
	user := models.User{}
	user.Name = userName
	err := o.Read(&user)
	if err != orm.ErrNoRows { //查询到已注册的用户
		logs.Info("当前用户信息为：", user)
		c.Data["userName"] = userName
		c.Data["errmsg"] = "当前用户名已存在，请重新输入！"
		c.TplName = "register.html"
		return
	} else {
		logs.Info("查询用户信息失败，err=", err)
		user.Pwd = pwd
		// 参数1：影响的函数；参数2：是否发生错误
		_, err = o.Insert(&user)
		if err != nil {
			logs.Info("注册失败，", err)
			// 重定向注册界面
			c.Redirect("/register", 302)
			return
		}
	}

	// 4、返回登录页面
	// 指定视图文件，同时能传递数据
	c.TplName = "login.html"
	// 重定向不能传递数据，但是速度快
	//c.Redirect("/login", 302)
}
