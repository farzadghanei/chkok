package chkok

import (
	"testing"
)

const UNAVAILABLE_USERNAME = "unavailable-user" // hopefully this won't be the uid on the running host


func TestGetUid(t *testing.T) {
	var err error
	uid, err := getUid("root")
	if err != nil {
		t.Errorf("get_uid root want no err got err %v", err)
	}
	if uid != 0 {
		t.Errorf("get_uid root want uid 0 err got %v", uid)
	}

	uid, err = getUid("10375")
	if err != nil {
		t.Errorf("get_uid 10375 want no err got err %v", err)
	}
	if uid != 10375 {
		t.Errorf("get_uid 10375 want uid got %v", uid)
	}

	uid, err = getUid("UNAVAILABLE_USERNAME")
	if err == nil {
		t.Errorf("get_uid UNAVAILABLE_USERNAME want err got no err")
	}
	if uid != 0 {
		t.Errorf("get_uid UNAVAILABLE_USERNAME want uid 0 err got %v", uid)
	}
}

func TestGetGid(t *testing.T) {
	var err error
	uid, err := getGid("root")
	if err != nil {
		t.Errorf("get_gid root want no err got err %v", err)
	}
	if uid != 0 {
		t.Errorf("get_gid root want uid 0 err got %v", uid)
	}

	uid, err = getGid("10375")
	if err != nil {
		t.Errorf("get_gid 10375 want no err got err %v", err)
	}
	if uid != 10375 {
		t.Errorf("get_gid 10375 want uid got %v", uid)
	}

	uid, err = getGid("UNAVAILABLE_GROUP")
	if err == nil {
		t.Errorf("get_gid UNAVAILABLE_GROUP want err got no err")
	}
	if uid != 0 {
		t.Errorf("get_gid UNAVAILABLE_GROUP want uid 0 err got %v", uid)
	}
}