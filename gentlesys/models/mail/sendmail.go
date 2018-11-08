package mail

//发送邮件的功能，目前暂时只支持126邮箱。理论上修改代码可以支持所有邮箱
import (
	"gentlesys/global"
	"strings"

	"github.com/astaxie/beego/logs"
	"github.com/go-gomail/gomail"
)

var senderMailAddr string
var senderMailPasswd string
var senderMailName string
var WebDomainName string

func init() {
	senderMailAddr = global.GetStringFromCfg("mail::senderMailAddr", "")
	//这个密码不是邮箱的登录密码，是126邮箱在账户里面设置的授权发送码
	senderMailPasswd = global.GetStringFromCfg("mail::senderMailPasswd", "")
	senderMailName = global.GetStringFromCfg("mail::senderMailName", "")
	WebDomainName = global.GetStringFromCfg("mail::webDomainName", "")
}

func SendMail(receMailAddr string, subject string, bodyinfo string) bool {
	r := strings.Index(receMailAddr, "@")
	if r <= 0 {
		return false
	}
	receiver := []byte(receMailAddr)[0:r]

	//logs.Error("发送给", string(receiver))

	m := gomail.NewMessage()

	m.SetAddressHeader("From", senderMailAddr /*"发件人地址"*/, senderMailName) // 发件人

	m.SetHeader("To", m.FormatAddress(receMailAddr, string(receiver))) // 收件人

	m.SetHeader("Subject", subject) // 主题

	m.SetBody("text/html", bodyinfo) // 正文

	d := gomail.NewPlainDialer("smtp.126.com", 465, senderMailAddr, senderMailPasswd) // 发送邮件服务器、端口、发件人账号、发件人密码
	if err := d.DialAndSend(m); err != nil {
		logs.Error("发送失败", receMailAddr, subject, err)
		return false
	}
	return true
}
