# go-rustdesk-server
Forked from [danbai225/go-rustdesk-server](https://github.com/danbai225/go-rustdesk-server).
> A lightweight, no-frills RustDesk server — up and running before your blood pressure has a chance to rise.
>
> Thinking of giving up and migrating to rsop project.

## ⬇️ Download
[GitHub Releases](https://github.com/mokanove/go-rustdesk-server/releases)
[Mirror by MoKanove](https://867678.xyz/doc/mirror)

## ✨ Features

- 🩺 **Designed to cure the high blood pressure caused by compiling the original RustDesk Server**
- 🚫 Docker support, api_server, and WebUI have been completely removed
- ⚡ Pure, lightweight, and ready to run out of the box
- 📦 OpenWrt support included (rsop project merged)

## ⚠️ Before You Start

- **TCP and UDP ports 21114–21119 must all be open**, or the server will not work correctly.
- Since this project has no WebUI, port `21114` and `21119` hosts a stub WebServer that always returns `200 OK` to prevent client-side errors.
- **OpenWrt users**: only `aarch64` and `amd64` CPU architectures are supported. Requires OpenWrt **25.12 or later** (apk package manager).

## 🚀 Quick Start

```bash
./go-rustdesk-server
```

That's it. The server will start listening on TCP/UDP ports **21114–21119**.  
Configure your RustDesk client just as you would with a standard rustdesk-server setup.

```bash
./go-rustdesk-server help
```
Using this to get help

## 🛠 OpenWrt Usage (rsop)

### Service Management
> Auto-start is usually enabled by default. If the service doesn't start automatically, run the `enable` command manually.

| Action | Command |
|--------|---------|
| Start | `/etc/init.d/rsop start` |
| Restart | `/etc/init.d/rsop restart` |
| Check status | `/etc/init.d/rsop status` |
| Enable on boot | `/etc/init.d/rsop enable` |
| Show Key | `cat /etc/rustdesk/id_ed25519.pub` |
| Doctor(check service) | `/etc/rustdesk/go-rustdesk-server doctor [IP/Domain]` |


## 🔨 Building from Source

### Binary

```bash
go build
```

### OpenWrt Package
[Generic Document](https://867678.xyz/doc/build)
> From the SDK root directory, run these additional steps before building:
>
> Replace `⚠️ARCH` and `⚠️LIBC` with the architecture and libc type of your target platform.
```bash
cd ⚠️sdk-root/package/rsop/root/etc/rustdesk
rm DONOTREMOVE
wget "https://github.com/mokanove/go-rustdesk-server/releases/latest/download/go-rustdesk-server-linux-⚠️ARCH-⚠️LIBC"
```

## ⚖️ License
> This project is licensed under the **[MoPL](https://github.com/mokanove/mokanove/blob/main/docs/license.md)**.
>
> Source repository using MIT <https://github.com/danbai225/go-rustdesk-server> so can change to MoPL.
>
> We copied: Rustdesk Server. It is licensed under the GNU AGPL Version 3 <https://www.gnu.org/licenses/agpl-3.0.html>.
