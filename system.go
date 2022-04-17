package chkok

import (
	"os/user"
	"strconv"
)

// getUid returns uid of the specified username, if user doesn't exist, but name is numeric, it's assumed the uid
func getUid(name string) (int, error) {
	var uid int
	var uidstr string
	user, err := user.Lookup(name)
	if err == nil {
		uidstr = user.Uid
	} else {
		uidstr = name
	}
	uid, err = strconv.Atoi(uidstr)
	return uid, err
}

// getGid returns gid of the specified group, if gropu doesn't exist, but name is numeric, it's assumed the gid
func getGid(name string) (int, error) {
	var gid int
	var gidstr string
	group, err := user.LookupGroup(name)
	if err == nil {
		gidstr = group.Gid
	} else {
		gidstr = name
	}
	gid, err = strconv.Atoi(gidstr)
	return gid, err
}