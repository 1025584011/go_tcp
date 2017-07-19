package config

import (
	"testing"
)

func TestConfigLoadFile(t *testing.T) {
	err := LoadFromFile("key.xml", "Key")
	if err != nil {
		t.Error("LoadFromFile", err)
		return
	}
	/*
		err = LoadFromNet("http://localhost:8000/config?name=test.xml", "ScenesServer")
		if err != nil {
			t.Error("LoadFromNet", err)
			//return
		}
		// */
	ListConfig()
	cfg := NewConfig()
	cfg.LoadFromFile("key.xml", "Key")
	cfg.LoadListFromFile("key.xml", "Key")
	cfg.ListConfig()

}
