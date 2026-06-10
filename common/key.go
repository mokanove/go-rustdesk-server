package common

import (
	"crypto/ed25519"
	"encoding/base64"
	"go-rustdesk-server/cmd"
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

	privText := base64.StdEncoding.EncodeToString(priv)
	pubText := base64.StdEncoding.EncodeToString(pub)

	if err := os.WriteFile(filepath.Join(dir, "id_ed25519"), []byte(privText), 0600); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "id_ed25519.pub"), []byte(pubText), 0644)
}

func LoadKey() {
	dir := exeDir()
	privPath := filepath.Join(dir, "id_ed25519")
	pubPath := filepath.Join(dir, "id_ed25519.pub")

	if !Exists(privPath) {
		if err := genKey(dir); err != nil {
			cmd.Fatal("gen key err:", err)
			return
		}
	}

	privText, err := os.ReadFile(privPath)
	if err != nil {
		cmd.Fatal("read priv key err:", err)
		return
	}
	pubText, err := os.ReadFile(pubPath)
	if err != nil {
		cmd.Fatal("read pub key err:", err)
		return
	}

	sk, err = base64.StdEncoding.DecodeString(string(privText))
	if err != nil {
		cmd.Fatal("decode priv key err:", err)
		return
	}

	pk, err = base64.StdEncoding.DecodeString(string(pubText))
	if err != nil {
		cmd.Fatal("decode pub key err:", err)
		return
	}

	pkStr = string(pubText)
	cmd.Info("key= %s", pkStr)
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