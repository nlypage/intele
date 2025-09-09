<div align="center">

# 🤖 Intele v2

**A powerful flow-based conversation framework for Telegram bots**

[![Go Reference](https://pkg.go.dev/badge/github.com/nlypage/intele/v2.svg)](https://pkg.go.dev/github.com/nlypage/intele/v2)
[![Go Report Card](https://goreportcard.com/badge/github.com/nlypage/intele/v2)](https://goreportcard.com/report/github.com/nlypage/intele/v2)
[![License](https://img.shields.io/github/license/nlypage/intele?color=blue)](LICENSE)

> ⚠️ **BETA WARNING**: This library is currently in beta development. It is NOT recommended for production use. APIs may change without notice, and breaking changes are expected until v2.0.0 stable release.

Built with [telebot.v3](https://github.com/tucnak/telebot) • Create complex multi-step conversations with session management, dependency injection, and middleware support.

</div>

---

## ✨ Features

<table>
<tr>
<td width="50%">

🔄 **Flow-based Architecture**  
Build complex multi-step conversations

💾 **Session Management**  
Persistent session state with automatic restoration

🏗️ **Builder Pattern**  
Fluent API for creating flows and steps

🔌 **Dependency Injection**  
Built-in DI container for step dependencies

🎯 **Middleware Support**  
Global and step-level middleware

</td>
<td width="50%">

📱 **Message Collection**  
Automatic message cleanup and management

⏱️ **Timeout Handling**  
Configurable session timeouts

🛡️ **Thread-safe**  
Concurrent session handling

🔀 **Dynamic Navigation**  
Jump between steps with validation

📊 **Multiple Input Types**  
Text, callbacks, contacts, locations, files

</td>
</tr>
</table>

## 🚀 Installation

```
go get github.com/nlypage/intele/v2
```

## 💡 Quick Start

```go
package main

import (
	"time"
        
	"github.com/nlypage/intele/v2"
	"github.com/nlypage/intele/v2/storage"
	tele "gopkg.in/telebot.v3"
)

func main() {
    bot, _ := tele.NewBot(tele.Settings{
    Token:  "YOUR_BOT_TOKEN",
    Poller: &tele.LongPoller{Timeout: 10 * time.Second},
    })

    // Create flow bus
    flowBus := intele.NewBus(bot, storage.NewMemoryStorage())
    
    // Setup handlers
    bot.Use(flowBus.Restore())
    bot.Handle(tele.OnText, flowBus.Handle)
    bot.Handle(tele.OnCallback, flowBus.Handle)

    // Create a simple flow
    flow, _ := flowBus.NewFlow("greeting").
        Steps(
            intele.NewStep("start", func(c tele.Context, ctrl intele.Controller) error {
                return c.Send("Hello! What's your name?")
            }).OnComplete(func(c tele.Context, ctrl intele.Controller) error {
                ctrl.Storage().Set("name", c.Text())
                return ctrl.Jump("finish")
            }),
            
            intele.NewStep("finish", func(c tele.Context, ctrl intele.Controller) error {
                name, _ := ctrl.Storage().GetString("name")
                _ = ctrl.Complete()
                return c.Send("Nice to meet you, " + name + "!")
            }),
        ).
        Build()

    bot.Handle("/start", flow.Start)
    bot.Start()
}
```

## 📖 Example

See the [`examples/`](examples/) directory for complete working examples

## 🤝 Contributing

Contributions are welcome! CONTRIBUTING.md will be added soon

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

- 🐛 [Issues](https://github.com/nlypage/intele/issues)
- 💬 [Discussions](https://github.com/nlypage/intele/discussions)

---

<div align="center">

**Made with ❤️ by [nlypage](https://github.com/nlypage)**

</div>