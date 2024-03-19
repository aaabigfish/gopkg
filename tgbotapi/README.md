# tgbotapi 

telegram库，基于 [tgbotapi](github.com/go-telegram-bot-api/telegram-bot-api/v5) 封装， 依赖gitlab.ipcloud.cc/go/gopkg/config 默认配置在项目conf/config.toml，配置模版如下：
```
[tgbot]
# 机器人token
TgBotToken = "6206174216:AAGud5WHErv6a7icyZw_fiMEii-2WWIVkZo"
# 通知的群id
TgChatId = -967311787

```


# 示例
```go
import "gitlab.ipcloud.cc/go/gopkg/tgbotapi"

var (
	telegramBotChatId1 int64 = -967311787
    telegramBotChatId2 int64 = -967311787
	telegramBotToken        = "6206174216:AAGud5WHErv6a7icyZw_fiMEii-2WWIVkZo"
)

// 创建发送实例,可以指定多个群id，也可以不指定
bot := tgbotapi.NewBot(telegramBotToken, telegramBotChatId1)

// 已经初始化了聊天id，可以使用SendMsg方法，不需要指定聊天
bot.SendMsg("testbot msg")

// 指定的聊天id发送
bot.Send(telegramBotChatId2, "testbot")

// 指定渠道发送
mq.SendChannelMsg("channelName", "test")

```


