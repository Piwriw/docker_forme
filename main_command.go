package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"github.piwriw.docker/container"
	"os"
	"strings"
)

var runCommand = cli.Command{
	Name: "run",
	Usage: `Create a container with namespace and cgroups limit
			mydocker run -it [command]`,
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "it", // 简单起见，这里把 -i 和 -t 参数合并成一个
			Usage: "enable tty",
		},
	},
	/*
		这里是run命令执行的真正函数。
		1.判断参数是否包含command
		2.获取用户指定的command
		3.调用Run function去准备启动容器:
	*/
	Action: func(context *cli.Context) error {
		if len(context.Args()) < 1 {
			return fmt.Errorf("missing container command")
		}
		cmd := context.Args().Get(0)
		tty := context.Bool("it")
		Run(tty, []string{cmd})
		return nil
	},
}

var initCommand = cli.Command{
	Name:  "init",
	Usage: "Init container process run user's process in container. Do not call it outside",
	/*
		1.获取传递过来的 command 参数
		2.执行容器初始化操作
	*/
	Action: func(context *cli.Context) error {
		log.Infof("init come on")
		cmd := context.Args().Get(0)
		log.Infof("command: %s", cmd)
		err := container.RunContainerInitProcess()
		return err
	},
}

// Run 执行具体 command
/*
这里的Start方法是真正开始前面创建好的command的调用，它首先会clone出来一个namespace隔离的
进程，然后在子进程中，调用/proc/self/exe,也就是调用自己，发送init参数，调用我们写的init方法，
去初始化容器的一些资源。
*/
func Run(tty bool, commandArr []string) {
	parent, writePipe := container.NewParentProcess(tty)
	if err := parent.Start(); err != nil {
		log.Error(err)
	}
	_ = parent.Wait()
	// 在子进程创建后通过管道来发送参数
	sendInitCommand(commandArr, writePipe)
	os.Exit(-1)
}

// sendInitCommand 通过writePipe将指令发送给子进程
func sendInitCommand(commandArr []string, writePipe *os.File) {
	command := strings.Join(commandArr, " ")
	log.Infof("command all is %s", command)
	_, _ = writePipe.WriteString(command)
	_ = writePipe.Close()
}
