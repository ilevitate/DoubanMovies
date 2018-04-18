package main

import (
	"net/http"
	"log"
	"io/ioutil"
	"regexp"
	"fmt"
	"strconv"
	"time"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

type Movie struct {
	Title string
	Score string
	Desc  string
	Url   string
	Cover string
}

//定义数据类型
type Spider struct {
	url    string
	header map[string]string
}

//定义Spider的 get 方法
func (keywords Spider) getHtmlHeader() string {
	//定义个一个http.client客户端
	client := &http.Client{}
	req, err := http.NewRequest("GET", keywords.url, nil)
	if err != nil {
		log.Panic("http.NewRequest GET url出错：", err.Error())
	}
	for key, value := range keywords.header {
		req.Header.Add(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Panic("client.Do req出错：", err.Error())
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Panic("ioutil.ReadAll 出错：", err.Error())
	}
	return string(body)
}

//获取数据
func parse() {
	header := map[string]string{
		"Host":                      "movie.douban.com",
		"Connection":                "keep-alive",
		"Cache-Control":             "max-age=0",
		"Upgrade-Insecure-Requests": "1",
		"User-Agent":                "Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/53.0.2785.143 Safari/537.36",
		"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
		"Referer":                   "https://movie.douban.com/top250",
	}

	////创建excel文件
	//f, err := os.Create("./douban.xls")
	//if err != nil {
	//	panic(err)
	//}
	//defer f.Close()
	////写入标题
	//f.WriteString("电影名称" + "\t" + "评分" + "\t" + "简介" + "\t" + "图片地址" + "\t" + "豆瓣链接" + "\t" + "\r\n")

	//循环每页解析并把结果写入excel插入数据库
	for i := 0; i < 10; i++ {
		fmt.Println("正在抓取第" + strconv.Itoa(i) + "页......")
		url := "https://movie.douban.com/top250?start=" + strconv.Itoa(i*25) + "&filter="
		spider := &Spider{url, header}
		html := spider.getHtmlHeader()

		//封面URL地址
		pattern0 := `" src="(.*?)" class="">`
		rp0 := regexp.MustCompile(pattern0)
		find_txt0 := rp0.FindAllStringSubmatch(html, -1)

		//豆瓣地址
		pattern1 := `<a href="(.*?)" class="">`
		rp1 := regexp.MustCompile(pattern1)
		find_txt1 := rp1.FindAllStringSubmatch(html, -1)

		//简介信息
		pattern2 := `<span class="inq">(.*?)</span>`
		rp2 := regexp.MustCompile(pattern2)
		find_txt2 := rp2.FindAllStringSubmatch(html, -1)

		//评分
		pattern3 := `<span class="rating_num" property="v:average">(.*?)</span>`
		rp3 := regexp.MustCompile(pattern3)
		find_txt3 := rp3.FindAllStringSubmatch(html, -1)

		//电影名称
		pattern4 := `<img width="100" alt="(.*?)" src="`
		rp4 := regexp.MustCompile(pattern4)
		find_txt4 := rp4.FindAllStringSubmatch(html, -1)


		//链接数据库
		db, err := sql.Open("mysql", "root:123456@/douban")
		if err != nil {
			log.Panic("链接数据库错误：", err.Error())
		}
		defer db.Close()
		err = db.Ping()
		if err != nil {
			panic(err.Error())
		}
		//// 写入UTF-8 BOM
		//f.WriteString("\xEF\xBB\xBF\xBB\xBB")
		//为什么是小于find_txt2的长度，因为有的影片没有简介。
		for i := 0; i < len(find_txt2); i++ {
			//fmt.Printf("%s %s %s %s %s\n",find_txt4[i][1],find_txt3[i][1],find_txt2[i][1], find_txt1[i][1],find_txt0[i][1])
			//f.WriteString(find_txt4[i][1] + "\t" + find_txt3[i][1] + "\t" + find_txt2[i][1] + "\t" + find_txt1[i][1] + "\t" + find_txt0[i][1] + "\r\n")
			movie := Movie{
				Title: find_txt4[i][1],
				Score: find_txt3[i][1],
				Desc:  find_txt2[i][1],
				Url:   find_txt1[i][1],
				Cover: find_txt0[i][1],
			}
			//入库操作
			stmt, err := db.Prepare("INSERT `movies` SET `title`=?,`score`=?,`desc`=?,`url`=?,`cover`=?")
			if err != nil {
				log.Panic("db.Prepare 数据库出错：", err.Error())
			}
			_, err = stmt.Exec(movie.Title, movie.Score, movie.Desc, movie.Url, movie.Cover)
			if err != nil {
				log.Panic("stmt.Exec 执行数据操作出错：", err.Error())
			}
		}

	}
}

func main() {
	log.SetFlags(log.Llongfile)
	t1 := time.Now() // get current time
	//调用方法执行爬虫
	parse()
	elapsed := time.Since(t1)
	fmt.Println("爬虫结束,总共耗时: ", elapsed)
}
