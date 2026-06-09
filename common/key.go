package common

import (
	"crypto/ed25519"
	"encoding/base64"
	logs "github.com/danbai225/go-logs"
	"go-rustdesk-server/model/model_proto"
	"google.golang.org/protobuf/proto"
	"os"
	"path/filepath"
)

var pk []byte
var pkStr string
var sk []byte

func exeDir() string {
	exe, err := os.Executable()
	if err != nil {
		return "."
	}
	return filepath.Dir(exe)
}

func genKey(dir string) error {
	pub, priv, _ := ed25519.GenerateKey(nil)
	if err := os.WriteFile(filepath.Join(dir, "id_ed25519"), priv, 0600); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "id_ed25519.pub"), pub, 0644)
}

func LoadKey() {
	dir := exeDir()
	privPath := filepath.Join(dir, "id_ed25519")
	pubPath := filepath.Join(dir, "id_ed25519.pub")
	if !Exists(privPath) {
		if err := genKey(dir); err != nil {
			logs.Err("gen key err:", err)
			return
		}
	}
	var err error
	sk, err = os.ReadFile(privPath)
	if err != nil {
		logs.Err("read key err:", err)
		return
	}
	pk, err = os.ReadFile(pubPath)
	if err != nil {
		logs.Err("read key err:", err)
		return
	}
	pkStr = base64.StdEncoding.EncodeToString(pk)
	logs.Info("key=", pkStr)
}

func Sign(data []byte) []byte {
	return append(ed25519.Sign(sk, data), data...)
}

func GetSignPK(version, id string, peerPK []byte) []byte {
	if version == "" || id == "" {
		return []byte{}
	}
	marshal, _ := proto.Marshal(&model_proto.IdPk{Id: id, Pk: peerPK})
	return Sign(marshal)
}

func GetPkStr() string { return pkStr }
