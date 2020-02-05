package controllers

import (
	"beegoDemo/models"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"time"
)

type LoginAndLogoutController struct {
	beego.Controller
}

func (c *LoginAndLogoutController) TurnToLogin() {
	userName := c.Ctx.GetCookie("userName")
	logs.Info("记住的用户名为：", userName)
	if userName != "" {
		c.Data["userName"] = userName
		c.Data["check"] = "checked"
	}
	logs.Info("当前跳转到登录页面！")
	c.TplName = "login.html"
}

func (c *LoginAndLogoutController) HandleLogin() {
	logs.Info("当前来到登录界面后台处理！")
	// 1、获取数据
	userName := c.GetString("userName")
	pwd := c.GetString("pwd")
	// 2、查询数据库
	o := orm.NewOrm()
	user := models.User{}
	user.Name = userName
	err := o.Read(&user, "Name")
	logs.Info("查询到当前用户信息为：", user)
	if err == orm.ErrNoRows {
		logs.Info("用户名不存在！")
		c.Data["userName"] = userName
		c.Data["errmsg"] = "当前用户名不存在！"
		c.TplName = "login.html"
		return
	}
	if user.Pwd != pwd {
		c.Data["userName"] = userName
		c.Data["errmsg"] = "登录密码错误，请重新输入！"
		c.TplName = "login.html"
		return
	}

	check := c.GetString("remember")
	logs.Info("是否记住用户名：", check)
	if check == "on" {
		//实现1小时内记住登录用户名，返回值为on
		logs.Info("设置了cookie！")
		c.Ctx.SetCookie("userName", userName, time.Second*3600)
	} else {
		// 设置登录cookie过期
		c.Ctx.SetCookie("userName", "sss", -1)
	}
	// 设置session
	c.SetSession("userName", userName)
	// 4、跳转页面
	// c.Ctx.WriteString("欢迎您，登录成功！")
	c.Redirect("/article/showArticleList", 302)
}

// 退出登录
func (c *LoginAndLogoutController) Logout() {
	// 1、删除登录状态
	c.DelSession("userName")
	// 回到登录界面
	c.Redirect("/", 302)
}
