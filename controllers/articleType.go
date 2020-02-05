package controllers

import (
	"beegoDemo/controllers/utils"
	"beegoDemo/models"
	"bytes"
	"encoding/gob"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"github.com/gomodule/redigo/redis"
)

type ArticleTypeController struct {
	beego.Controller
}

// 跳转到添加文章类型的页面
func (c *ArticleTypeController) TurnToAddArticleType() {
	// 1、读取类型表，显示数据
	var articleTypes []models.ArticleType

	conn := utils.Pool.Get()
	defer conn.Close()
	buf, _ := redis.String(conn.Do("get", "types"))
	dec := gob.NewDecoder(bytes.NewReader([]byte(buf)))
	err := dec.Decode(&articleTypes)
	logs.Info("解码后的类型数组为：", articleTypes)
	if err != nil {
		logs.Info("解码失败，err=", err)
	}

	c.Data["types"] = articleTypes
	c.Layout = "layout.html"
	c.LayoutSections = make(map[string]string)
	c.LayoutSections["headTitle"] = "title/addArticleTypeTitle.html"
	c.Data["userName"] = c.GetSession("userName")
	c.TplName = "addArticleType.html"

}

// 处理添加类型业务
func (c *ArticleTypeController) HandleAddArticleType() {
	typename := c.GetString("typeName")
	if typename == "" {
		logs.Info("添加类型数据为空！")
		return
	}
	o := orm.NewOrm()
	var articleType models.ArticleType
	articleType.Typename = typename
	_, err := o.Insert(&articleType)
	if err != nil {
		logs.Info("添加分类失败，err=", err)
		return
	}

	var types [] models.ArticleType
	// 去数据库查询
	_, err = o.QueryTable("ArticleType").All(&types)
	if err != nil {
		logs.Info("查询分类内容失败，err=", err)
		return
	}
	// 更新redis中的一个键值对
	conn := utils.Pool.Get()
	defer conn.Close()
	// 字节缓冲区
	var buffer bytes.Buffer
	// 编码
	enc := gob.NewEncoder(&buffer)
	_ = enc.Encode(&types)
	// 键对应的值存放buffer的字节切片，即buffer.Bytes()
	_, _ = conn.Do("set", "types", buffer.Bytes())
	// 展示视图
	c.Redirect("/article/addArticleType", 302)
}

// 删除文章类型，默认是级联删除
func (c *ArticleTypeController) HandleDeleteArticleType() {
	id, _ := c.GetInt("id")
	o := orm.NewOrm()
	Atype := models.ArticleType{Id: id}
	_, err := o.Delete(&Atype)
	if err != nil {
		logs.Info("删除失败，err=", err)
	}

	// 去数据库查询所有类型
	var types [] models.ArticleType
	_, err = o.QueryTable("ArticleType").All(&types)
	if err != nil {
		logs.Info("查询分类内容失败，err=", err)
		return
	}
	// 更新redis中的一个键值对
	conn := utils.Pool.Get()
	defer conn.Close()
	// 字节缓冲区
	var buffer bytes.Buffer
	// 编码
	enc := gob.NewEncoder(&buffer)
	_ = enc.Encode(&types)
	// 键对应的值存放buffer的字节切片，即buffer.Bytes()
	_, _ = conn.Do("set", "types", buffer.Bytes())

	// 重定向，否则地址栏不变会发生错误
	c.Redirect("/article/addArticleType", 302)
}
