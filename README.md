# go-rustdesk-server
Forked from：<https://github.com/danbai225/go-rustdesk-server>
## ⬇️ Downloads
[GitHub Release](https://github.com/mokanove/go-rustdesk-server/releases)
## 🚀 Features
- **Designed to cure the high blood pressure and heart attacks caused by compiling the original RustDesk Server.**
- Completely stripped of Docker support, api_server, and WebUI.
- **Functions as a pure, lightweight RustDesk Server.**
## ⚠️ Warning
> You must ensure that both TCP and UDP ports between 21114 and 21119 are open; otherwise, the server will not function properly.
>
> Since this project does not require a WebUI, port 21114 hosts a fake WebServer that always responds 200 OK to prevent client errors.
## 🧰 How to use
```
./go-rustdesk-server
```
> Yes, it's that simple!
>
> If there are no problems, he should start listening on TCP and UDP ports between 21114 and 21119.
>
> Then you can fill in the information just like you would when configuring a normal rustdesk-server.
## 🛠 How to self-build
```
go build
```
> Yes , so simple! But you need Golang Version 1.26.4 or later.
>
> This is the biggest reason why I rewrote it in Go.
## ⚖️ License
> This project is licensed under the **[MoPL](https://github.com/mokanove/mokanove/blob/main/docs/license.md)**
>
> The original project uses MIT, so it can be replaced with MoPL.
