package main

import (
	"github.com/gin-gonic/gin"
	"github.com/guotie/deferinit"
	"github.com/smtc/glog"
	"github.com/swgloomy/gutil"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
)

var (
	port    = ":1201"
	pidPath = "./httpPrint.pid" //pid文件
	logDir  = "./logs"
	rt      *gin.Engine
)

func main() {
	//if gutil.CheckPid(pidPath) {
	//	return
	//}
	gutil.LogInit(true, logDir)
	serverStart()

	rt = gin.Default()
	rt.Use(Cors())
	router(rt)

	go func() {
		err := rt.Run(port)
		if err != nil {
			glog.Error("main gin run err! err: %s \n", err.Error())
		}
	}()

	c := make(chan os.Signal, 1)
	//gutil.WritePid(pidPath)
	// 信号处理
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM)
	// 等待信号
	<-c
	serverExit()
	//gutil.RmPidFile(pidPath)
	os.Exit(0)
}

func router(r *gin.Engine) {
	g := &r.RouterGroup
	g.GET("/", func(c *gin.Context) { c.String(http.StatusOK, "ok") })
	g.GET("/Print", printPdf)
}

func printPdf(c *gin.Context) {
	fileUrl := c.Query("file")
	go fileDown(fileUrl)
	c.String(http.StatusOK, "打印文件处理中!")
}

func fileDown(fileUrl string) {
	resp, err := http.Get(fileUrl)
	if err != nil {
		glog.Error("router http downLoad,err: %s \n", err)
		return
	}
	defer func() {
		err = resp.Body.Close()
		if err != nil {
			glog.Error("fileDown resp body close err! err: %s \n", err.Error())
			return
		}
	}()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		glog.Error("fileDown http read err! err: %s \n", err.Error())
		return
	}
	pdfPath := "./gloomy.pdf"
	bo, err := gutil.PathExists(pdfPath)
	if err != nil {
		glog.Error("fileDown path exists file run err! err: %s \n", err.Error())
	}
	if bo {
		err = os.Remove(pdfPath)
		if err != nil {
			glog.Error("fileDown remove file err! err: %s \n", err.Error())
			return
		}
	}
	err = ioutil.WriteFile(pdfPath, data, 0644)
	if err != nil {
		glog.Error("fileDown write file err! err: %s \n", err.Error())
		return
	}
	cmd := exec.Command("./SumatraPDF.exe", "-print-to-default", pdfPath)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		glog.Error("fileDown StdoutPipe run err! err: %s \n", err.Error())
		return
	}
	defer func() {
		err = stdout.Close()
		if err != nil {
			glog.Error("fileDown stdout close err! err: %s \n", err.Error())
			return
		}
	}()

	err = cmd.Start()
	if err != nil {
		glog.Error("fileDown cmd start err! err: %s \n", err.Error())
		return
	}
	opBytes, err := ioutil.ReadAll(stdout)
	if err != nil {
		glog.Error("fileDown cmd read result err! err: %s \n", err.Error())
		return
	}
	glog.Info("fileDown cmd start success! result: %s \n", string(opBytes))
}

func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method

		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "Content-Type,AccessToken,X-CSRF-Token, Authorization, Token")
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
		c.Header("Access-Control-Allow-Credentials", "true")

		//放行所有OPTIONS方法
		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}
		// 处理请求
		c.Next()
	}
}

func serverStart() {
	//初始化所有go文件中的init内方法
	deferinit.InitAll()
	glog.Info("init all module successfully \n")

	//设置多CPU运行
	runtime.GOMAXPROCS(runtime.NumCPU())
	glog.Info("set many cpu successfully \n")

	//启动所有go文件中的init方法
	deferinit.RunRoutines()
	glog.Info("init all run successfully \n")
}

func serverExit() {
	// 结束所有go routine
	deferinit.StopRoutines()
	glog.Info("stop routine successfully.\n")
	deferinit.FiniAll()
	glog.Info("fini all modules successfully.\n")

	glog.Close()
}
