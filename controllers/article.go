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
	"math"
	"path"
	"strconv"
	"time"
)

type ArticleController struct {
	beego.Controller
}

// 显示文章列表页
func (c *ArticleController) TurnToArticleList() {
	/*	// 获取session
		userName := c.GetSession("userName") // 返回值为interface{}
		logs.Info("登录成功！")
		if userName == nil {
			// 登录拦截
			c.Redirect("/", 302)
			return
		}*/
	// 创建orm对象
	o := orm.NewOrm()
	// 定义切片保存文章list
	var articles []models.Article
	// 获取页码，默认为首页
	pageIndex, err := c.GetInt("pageIndex")
	if err != nil {
		pageIndex = 1
	}

	// 查询所有符合条件的文章
	// 每个页面显示的文章个数pageSize
	pageSize := 3

	var total, pageNum int64
	// 获取别名
	typeName := c.GetString("select")
	logs.Info("下拉框选中的文章别名为：", typeName)
	// 通过 User__Name 反向查询 Article
	// 加上RelatedSel()避免懒加载
	qs := o.QueryTable("Article").RelatedSel().Filter("User__Name", c.GetSession("userName"))
	if typeName == "" { // 默认查询所有
		typeName = c.GetString("typeName")
		if typeName == "" {
			logs.Info("下拉框传递数据失败！")
			// 没有类型名就查询所有
			total, _ = qs.Count()
			logs.Info("查询文章的总记录数：", total)
			_, err = qs.Limit(pageSize, pageSize*(pageIndex-1)).All(&articles)
			if err != nil {
				logs.Info("查询所有文章信息失败，err=", err)
				return
			}
		} else {
			logs.Info("当前默认文章类型为：", typeName)
			qs = qs.Filter("Atype__Typename", typeName)
			total, _ = qs.Count()
			_, err = qs.Limit(pageSize, pageSize*(pageIndex-1)).All(&articles)
			logs.Info("查询的总记录数：", total)
			if err != nil {
				logs.Info("查询所有文章信息失败，", err)
				return
			}
		}
		// return
	} else { // 根据类型字段来查询
		logs.Info("选中的文章类型为：", typeName)
		qs = qs.Filter("Atype__Typename", typeName)
		total, _ = qs.Count()
		_, err = qs.Limit(pageSize, pageSize*(pageIndex-1)).All(&articles)
		logs.Info("查询的总记录数：", total)
		if err != nil {
			logs.Info("查询所有文章信息失败，", err)
			return
		}
	}

	// 总页数
	pageNum = int64(math.Ceil(float64(total) / float64(pageSize)))

	// Limit()函数，参数1：一个页面显示的文章条数，参数2：偏移量，下标从0开始
	// 分页查询，RelatedSel()关联哪张表，左连接，多对一查询

	// 获取分类的内容
	var types [] models.ArticleType

	// 先去redis中查询是否有数据，若有则直接取出，否则再去数据库中查询
	// 从redis连接池中取出一个连接
	conn := utils.Pool.Get()
	defer conn.Close()
	buf, err1 := redis.String(conn.Do("get", "types"))
	if err1 != nil {
		logs.Info("查询失败,err=", err1)
		// 去数据库查询
		_, err1 = o.QueryTable("ArticleType").All(&types)
		if err1 != nil {
			logs.Info("查询分类内容失败，err=", err1)
			return
		}
		// 字节缓冲区
		var buffer bytes.Buffer
		// 编码
		enc := gob.NewEncoder(&buffer)
		_ = enc.Encode(&types)
		// 键对应的值存放buffer的字节切片，即buffer.Bytes()
		_, _ = conn.Do("set", "types", buffer.Bytes())
	} else {
		// logs.Info("数据为：", buf)
		// 解码器，构造一个NewReader再将buf转为字节
		dec := gob.NewDecoder(bytes.NewReader([]byte(buf)))
		err1 = dec.Decode(&types)
		logs.Info("解码后的类型数组为：", types)
		if err1 != nil {
			logs.Info("解码失败，err=", err1)
		}
	}
	c.Data["types"] = types
	c.Data["tname"] = typeName
	c.Data["total"] = total
	c.Data["articles"] = articles
	c.Data["pageNum"] = pageNum
	c.Data["pageIndex"] = pageIndex
	c.LayoutSections = make(map[string]string)
	c.LayoutSections["headTitle"] = "title/articleListTitle.html"
	c.Data["userName"] = c.GetSession("userName")
	c.TplName = "articleList.html"
}

// 显示添加文章页面
func (c *ArticleController) TurnToAddArticle() {
	// 查询类型数据，显示到添加文章页面中
	var types [] models.ArticleType
	// 从连接池中取出一个连接
	conn := utils.Pool.Get()
	defer conn.Close()
	buf, _ := redis.String(conn.Do("get", "types"))
	dec := gob.NewDecoder(bytes.NewReader([]byte(buf)))
	err := dec.Decode(&types)
	logs.Info("解码后的类型数组为：", types)
	if err != nil {
		logs.Info("解码失败，err=", err)
	}
	c.Data["types"] = types
	c.Layout = "layout.html"
	c.LayoutSections = make(map[string]string)
	c.LayoutSections["headTitle"] = "title/addArticleTitle.html"
	c.Data["userName"] = c.GetSession("userName")
	c.TplName = "addArticle.html"
}

// 处理添加文章界面数据
func (c *ArticleController) HandleAddArticle() {
	// 1、拿到数据
	articleName := c.GetString("articleName")
	articleContent := c.GetString("content")
	// logs.Info(articleName, articleContent)
	// 2、判断数据是否合法
	if articleName == "" || articleContent == "" {
		logs.Info("添加数据不能为空！")
		return
	}
	f, h, err := c.GetFile("uploadname")
	if err != nil {
		logs.Info("上传文件失败，", err)
		return
	}
	defer f.Close()
	// 要限定文件格式
	// 获取文件后缀
	fileExt := path.Ext(h.Filename)
	if fileExt != ".jpg" && fileExt != ".png" {
		logs.Info("上传文件格式错误！")
		return
	}
	// logs.Info(fileExt)
	// 限制文件大小（单位：字节）
	if h.Size > 1024*1024*10 {
		logs.Info("上传文件过大！")
		return
	}
	// 对文件进心重命名，防止文件名重复，用时间戳来保存
	fileName := strconv.FormatInt(time.Now().Unix(), 10) + fileExt // 6-1-2 15:4:5 为go语言的诞生时间

	// 注意图片地址路径前面要加点
	_ = c.SaveToFile("uploadname", "./static/img/"+fileName)
	// 3、插入数据
	o := orm.NewOrm()
	arti := models.Article{}
	arti.Aname = articleName
	arti.Acontent = articleContent
	// 这里图片路径不用加点
	arti.Aimg = "/static/img/" + fileName

	// 获取到下拉选项值
	typeName := c.GetString("select")
	if typeName == "" {
		logs.Info("下拉框数据为空！")
		return
	}
	// 获取type对象
	var artiType models.ArticleType
	artiType.Typename = typeName
	err = o.Read(&artiType, "Typename")
	if err != nil {
		logs.Info("获取类型错误，err=", err)
		return
	}
	// 传递的是一个指针类型
	arti.Atype = &artiType
	var user models.User
	user.Name = c.GetSession("userName").(string)
	err = o.Read(&user, "Name")
	if err != nil {
		logs.Info("获取用户数据失败, err=", err)
		return
	}
	arti.User = &user
	// 插入数据
	_, err = o.Insert(&arti)
	if err != nil {
		logs.Info("插入数据失败！")
		return
	}
	// 4、返回文章界面
	c.Redirect("/article/showArticleList", 302)
}

// 显示文章详情页面
func (c *ArticleController) TurnToArticleDetail() {
	// 1、获取文章id
	id, err := c.GetInt("id")
	if err != nil {
		logs.Info("获取文章Id错误，err=", err)
		return
	}
	// 2、查询数据库获取数据
	o := orm.NewOrm()
	arti := models.Article{Id: id}
	err = o.Read(&arti)
	if err != nil {
		logs.Info("查询数据为空,err=", err)
		return
	}
	arti.Acount++
	// 文章阅读量加1
	_, err = o.Update(&arti, "Acount")
	if err != nil {
		logs.Info("更新文章计数失败，err=", err)
		return
	}
	err = o.QueryTable("Article").RelatedSel().Filter("id", id).One(&arti)
	/*arti.Id = id
	  err = o.Read(&arti)*/
	if err != nil {
		logs.Info("查询文章失败，err=", err)
		return
	}
	// 3、传递数据给视图
	c.Data["article"] = arti
	logs.Info("当前文章内容为：", arti, "文章类型为：", arti.Atype.Typename)
	c.Layout = "layout.html"
	c.LayoutSections = make(map[string]string)
	c.LayoutSections["headTitle"] = "title/articleDetailTitle.html"
	c.Data["userName"] = c.GetSession("userName")
	c.TplName = "articleDetail.html"
}

// 显示编辑页面
func (c *ArticleController) TurnToUpdateArticle() {
	// 1、获取文章id
	id, err := c.GetInt("id")
	if err != nil {
		logs.Info("获取文章Id错误，err=", err)
		return
	}
	// 2、查询数据库获取数据
	o := orm.NewOrm()
	arti := models.Article{}
	arti.Id = id
	err = o.Read(&arti)
	if err != nil {
		logs.Info("查询文章失败，err=", err)
		return
	}
	// 3、传递数据给视图
	c.Data["article"] = arti
	c.Layout = "layout.html"
	c.LayoutSections = make(map[string]string)
	c.LayoutSections["headTitle"] = "title/updateArticleTitle.html"
	c.Data["userName"] = c.GetSession("userName")
	c.TplName = "updateArticle.html"
}

// 处理更新业务数据
func (c *ArticleController) HandleUpdateArticle() {
	// 1、拿到数据
	id, err1 := c.GetInt("id")
	if err1 != nil {
		logs.Info("获取文章Id错误，err=", err1)
		return
	}
	articleName := c.GetString("articleName")
	articleContent := c.GetString("content")
	// logs.Info(articleName, articleContent)
	// 2、判断数据是否合法
	if articleName == "" || articleContent == "" {
		logs.Info("更新字段不能为空！")
		return
	}
	f, h, err := c.GetFile("uploadname")
	var isUploadFile = true
	var fileName string
	if err != nil {
		if f != nil {
			logs.Info("上传文件失败，", err)
			return
		} else {
			logs.Info("没有上传图片，", err)
			// 相应的原先图片路径要保存
			isUploadFile = false
		}
	} else {
		defer f.Close()
		// 限定文件格式
		// 获取文件后缀
		fileExt := path.Ext(h.Filename)
		if fileExt != ".jpg" && fileExt != ".png" {
			logs.Info("上传文件格式错误！")
			return
		}
		// logs.Info(fileExt)
		// 限制文件大小（单位：字节）
		if h.Size > 1024*1024 {
			logs.Info("上传文件过大！")
			return
		}
		// 对文件进心重命名，防止文件名重复，用时间戳来保存
		fileName = strconv.FormatInt(time.Now().Unix(), 10) + fileExt // 6-1-2 3:4:5 为go语言的诞生时间

		// 注意图片地址路径前面要加点
		_ = c.SaveToFile("uploadname", "./static/img/"+fileName)
	}

	// 3、更新操作
	o := orm.NewOrm()
	arti := models.Article{Id: id}
	// 要先去查询数据，再更新数据
	err = o.Read(&arti)
	if err != nil {
		logs.Info("查询数据错误，err=", err)
		return
	}
	arti.Aname = articleName
	arti.Acontent = articleContent
	if isUploadFile {
		arti.Aimg = "/static/img/" + fileName
		// 指定更新的字段
		_, err = o.Update(&arti, "Aname", "Acontent", "Aimg")
	} else {
		_, err = o.Update(&arti, "Aname", "Acontent")
	}
	if err != nil {
		logs.Info("更新文章数据失败，err=", err)
		return
	}
	// 4、返回列表页面，重定向返回列表
	c.Redirect("/article/showArticleList", 302)
}

func (c *ArticleController) HandleDeleteArticle() {
	// 1、拿到数据
	id, err := c.GetInt("id")
	if err != nil {
		logs.Info("获取文章Id错误，err=", err)
		return
	}
	// 2、执行删除操作
	o := orm.NewOrm()
	arti := models.Article{}
	arti.Id = id
	err = o.Read(&arti)
	if err != nil {
		logs.Info("查询数据错误，err=", err)
		return
	}
	_, err = o.Delete(&arti)
	if err != nil {
		logs.Info("删除文章失败，err=", err)
		return
	}
	// 4、返回列表页面
	c.Redirect("/article/showArticleList", 302)
}
