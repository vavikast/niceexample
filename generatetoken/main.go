package main

import (
	"flag"
	"fmt"
	"github.com/skip2/go-qrcode"
	"github.com/vavikast/gotp"
	"os/exec"
	"runtime"
)

//目的
//生成谷歌密码验证器
//打印账号密码
//打印而密码qr出来

//主要使用github.com/xlzd/gotp
//但是它将特殊字符转义了，所以稍微做了修改，让它支持带@的用户名

var (
	user = flag.String("u", "", "请输入用户名(英文)，")
)
var (
	IssuerName  string //认证机构
	AccountName string //账号密码
)

func main() {
	//设置输入账号
	flag.Parse()
	//通过命令行输入src地址和目的地址
	if *user == "" {
		fmt.Println("请输入用户名(英文)，example: gotp -c g:")
		return
	}

	//设定签发者名称issuerName
	IssuerName = "Felix"

	//设置用户账户和认证秘钥密码，随机生成密码
	AccountName = *user

	//存入本地存储，打印输出账号和密码，此处设置为32bit
	Secret := gotp.RandomSecret(32)

	//将账号和密码存入后台系统
	//do.insert(db)

	//输入账号和密码信息
	fmt.Printf("你的签发机构是：%v,账号是：%v, 生成随机secret是：%v,请及时保存！\n", IssuerName, AccountName, Secret)

	//打印输出二维码
	qrurl := gotp.NewDefaultTOTP(Secret).ProvisioningUri(AccountName, IssuerName)

	qrcode.WriteFile(qrurl, qrcode.Medium, 256, "./qrcode.png")

	//如果是windows系统打开二维码
	if runtime.GOOS == "windows" {
		fmt.Println("Android手机APK下载: https://os-android.liqucn.com/rj/225046.shtml iOS手机直接:AppStore 搜索 google authenticator")
		exec.Command("cmd", "/c", "start ./qrcode.png")
	}

}
